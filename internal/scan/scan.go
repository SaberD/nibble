package scan

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scanner"
)

// NetScanner performs real network scanning (TCP connect, ARP, banner grab).
type NetScanner struct {
	Ports []int
}

type portResult struct {
	port   int
	banner string
}

// ScanNetwork scans a real subnet with controlled concurrency for smooth progress.
func (s *NetScanner) ScanNetwork(ifaceName, subnet string, progressChan chan<- scanner.ProgressUpdate) {
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return
	}

	totalHosts := scanner.TotalScanHosts(ipnet)

	skipIPs := s.runNeighborDiscoveryPhase(ifaceName, ipnet, totalHosts, progressChan)
	s.runSubnetSweepPhase(ifaceName, ipnet, totalHosts, skipIPs, progressChan)
	close(progressChan) // Signal completion
}

// runNeighborDiscoveryPhase emits hosts already visible in neighbor tables
// and returns IPs that should be skipped in the full sweep.
func (s *NetScanner) runNeighborDiscoveryPhase(ifaceName string, subnet *net.IPNet, totalHosts int, progressChan chan<- scanner.ProgressUpdate) map[string]struct{} {
	discovered := visibleNeighbors(ifaceName, subnet)
	ports := s.ports()
	skipIPs := make(map[string]struct{}, len(discovered))
	for _, neighbor := range discovered {
		skipIPs[neighbor.IP] = struct{}{}
	}

	if len(discovered) == 0 {
		select {
		case progressChan <- scanner.NeighborProgress{
			TotalHosts: totalHosts,
			Seen:       0,
			Total:      0,
		}:
		default:
		}
		return skipIPs
	}

	workers := 16
	if len(discovered) < workers {
		workers = len(discovered)
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, workers)
	var progressMu sync.Mutex
	phaseDone := 0

	for _, neighbor := range discovered {
		wg.Add(1)
		sem <- struct{}{}
		go func(neighbor NeighborEntry) {
			defer wg.Done()
			defer func() { <-sem }()

			hostInfo := scanHostWithKnownMAC(ifaceName, neighbor.IP, neighbor.MAC, ports)
			if hostInfo == "" {
				hardware := vendorFromMac(neighbor.MAC)
				hostInfo = scanner.FormatHost(scanner.HostResult{
					IP:       neighbor.IP,
					Hardware: hardware,
				})
			}

			progressMu.Lock()
			phaseDone++
			currentDone := phaseDone
			progressMu.Unlock()

			select {
			case progressChan <- scanner.NeighborProgress{
				Host:       hostInfo,
				TotalHosts: totalHosts,
				Seen:       currentDone,
				Total:      len(discovered),
			}:
			default:
			}
		}(neighbor)
	}
	wg.Wait()

	return skipIPs
}

func scanOpenPorts(ip string, ports []int) []portResult {
	var wg sync.WaitGroup
	var mutex sync.Mutex
	var results []portResult

	// Scan all ports in parallel for this host
	for _, port := range ports {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()

			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), 200*time.Millisecond)
			if err != nil {
				return
			}

			banner := grabServiceBanner(conn, port)
			conn.Close()

			mutex.Lock()
			results = append(results, portResult{port: port, banner: banner})
			mutex.Unlock()
		}(port)
	}

	wg.Wait()
	return results
}

func resolveHardware(ifaceName string, targetIP net.IP, knownMAC string) string {
	if knownMAC != "" {
		return vendorFromMac(knownMAC)
	}

	if targetIP == nil {
		return ""
	}

	mac := resolveMac(ifaceName, targetIP)
	if mac == "" {
		mac = lookupMacFromCache(targetIP.String())
	}
	if mac == "" {
		return ""
	}
	return vendorFromMac(mac)
}

func scanHostWithKnownMAC(ifaceName string, ip string, knownMAC string, ports []int) string {
	results := scanOpenPorts(ip, ports)
	if len(results) == 0 {
		return ""
	}

	// Resolve hardware vendor from known MAC, ARP, or cache.
	hardware := resolveHardware(ifaceName, net.ParseIP(ip), knownMAC)

	// Sort by port number
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].port < results[i].port {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Build output using shared formatter
	result := scanner.HostResult{IP: ip, Hardware: hardware}
	for _, r := range results {
		result.Ports = append(result.Ports, scanner.PortInfo{Port: r.port, Banner: r.banner})
	}

	return scanner.FormatHost(result)
}

// runSubnetSweepPhase scans the subnet and skips hosts found in neighbor discovery.
func (s *NetScanner) runSubnetSweepPhase(ifaceName string, subnet *net.IPNet, totalHosts int, skipIPs map[string]struct{}, progressChan chan<- scanner.ProgressUpdate) {
	var wg sync.WaitGroup
	var mutex sync.Mutex
	scanned := 0
	ports := s.ports()

	// Semaphore to limit concurrent host scans (100 for speed)
	semaphore := make(chan struct{}, 100)

	incrementIP := func(ip net.IP) {
		for j := len(ip) - 1; j >= 0; j-- {
			ip[j]++
			if ip[j] > 0 {
				break
			}
		}
	}

	// Scan hosts with controlled concurrency
	for ip := subnet.IP.Mask(subnet.Mask); subnet.Contains(ip); incrementIP(ip) {
		// Skip subnet network address and IPv4 broadcast address.
		if isSubnetNetworkOrBroadcastIPv4(ip, subnet) {
			continue
		}

		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore

		go func(currentIP string) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore

			hostInfo := ""
			if _, alreadyFound := skipIPs[currentIP]; !alreadyFound {
				hostInfo = scanHost(ifaceName, currentIP, ports)
			}

			mutex.Lock()
			scanned++
			currentScanned := scanned
			mutex.Unlock()

			// Send progress update for every host
			select {
			case progressChan <- scanner.SweepProgress{
				Host:       hostInfo,
				TotalHosts: totalHosts,
				Scanned:    currentScanned,
				Total:      totalHosts,
			}:
			}
		}(ip.String())
	}

	wg.Wait()
}

func isSubnetNetworkOrBroadcastIPv4(ip net.IP, subnet *net.IPNet) bool {
	ip4 := ip.To4()
	base := subnet.IP.To4()
	mask := subnet.Mask
	if ip4 == nil || base == nil || len(mask) != net.IPv4len {
		return false
	}

	// Network address: host bits all zero.
	if ip4.Equal(base.Mask(mask)) {
		return true
	}

	// Broadcast address: host bits all one.
	broadcast := net.IPv4(
		base[0]|^mask[0],
		base[1]|^mask[1],
		base[2]|^mask[2],
		base[3]|^mask[3],
	)
	return ip4.Equal(broadcast)
}

// scanHost resolves hardware via ARP, scans ports, grabs banners.
func scanHost(ifaceName string, ip string, ports []int) string {
	return scanHostWithKnownMAC(ifaceName, ip, "", ports)
}

func (s *NetScanner) ports() []int {
	if len(s.Ports) > 0 {
		return s.Ports
	}
	return ports.DefaultPorts()
}

// grabServiceBanner reads a service banner.
// Push-banner protocols are read directly; HTTP ports send HEAD first.
func grabServiceBanner(conn net.Conn, port int) string {
	conn.SetDeadline(time.Now().Add(300 * time.Millisecond))

	switch port {
	case 80, 443, 8080, 8000, 8443:
		// HTTP - must send a request to get a response
		fmt.Fprintf(conn, "HEAD / HTTP/1.0\r\nHost: %s\r\n\r\n", conn.RemoteAddr())
		buf := make([]byte, 2048)
		n, err := conn.Read(buf)
		if err != nil || n == 0 {
			return ""
		}
		return parseHTTPServer(string(buf[:n]))

	default:
		// Push-banner protocols (SSH, FTP, SMTP, Telnet, etc.)
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil || n == 0 {
			return ""
		}
		response := sanitizeBannerBytes(buf[:n])
		// Clean up newlines for display
		response = strings.ReplaceAll(response, "\r\n", " ")
		response = strings.ReplaceAll(response, "\n", " ")
		if len(response) > 80 {
			response = response[:80]
		}
		return response
	}
}

func sanitizeBannerBytes(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}

	// Keep only first protocol line; avoids binary payload fragments (e.g. SSH KEX).
	if idx := bytesIndexAny(raw, '\n', '\r'); idx >= 0 {
		raw = raw[:idx]
	}

	// Replace invalid UTF-8 with '.', then keep printable ASCII-ish text.
	s := string(raw)
	if !utf8.ValidString(s) {
		var cleaned []rune
		for len(raw) > 0 {
			r, size := utf8.DecodeRune(raw)
			if r == utf8.RuneError && size == 1 {
				cleaned = append(cleaned, '.')
				raw = raw[1:]
				continue
			}
			cleaned = append(cleaned, r)
			raw = raw[size:]
		}
		s = string(cleaned)
	}

	var out []rune
	for _, r := range strings.TrimSpace(s) {
		if r >= 32 && r <= 126 {
			out = append(out, r)
		}
	}
	return strings.TrimSpace(string(out))
}

func bytesIndexAny(b []byte, chars ...byte) int {
	for i, c := range b {
		for _, ch := range chars {
			if c == ch {
				return i
			}
		}
	}
	return -1
}

// parseHTTPServer extracts the Server header from an HTTP response
func parseHTTPServer(response string) string {
	for _, line := range strings.Split(response, "\r\n") {
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "server:") {
			return normalizeBannerForDisplay(strings.TrimSpace(line[7:]))
		}
	}
	// No Server header, try to return the status line
	if idx := strings.Index(response, "\r\n"); idx > 0 {
		return normalizeBannerForDisplay(response[:idx])
	}
	return normalizeBannerForDisplay(response)
}

func normalizeBannerForDisplay(s string) string {
	clean := sanitizeBannerBytes([]byte(s))
	if len(clean) > 80 {
		clean = clean[:80]
	}
	return clean
}
