package scan

import (
	"fmt"
	"net"
	"unicode/utf8"
	"strings"
	"sync"
	"time"
)

// CommonPorts are the most common ports to scan for host discovery
var CommonPorts = []int{
	22,   // SSH
	23,   // Telnet
	80,   // HTTP
	443,  // HTTPS
	445,  // SMB
	3389, // RDP
	8080, // HTTP-Alt
}

// PortInfo holds a port number and its service banner
type PortInfo struct {
	Port   int
	Banner string
}

// HostResult holds all scan info for a single host
type HostResult struct {
	IP       string
	Hardware string
	Ports    []PortInfo
}

// FormatHost renders a HostResult into the display string
func FormatHost(h HostResult) string {
	var lines []string
	if h.Hardware != "" {
		lines = append(lines, fmt.Sprintf("%s - %s", h.IP, h.Hardware))
	} else {
		lines = append(lines, h.IP)
	}
	for _, p := range h.Ports {
		if p.Banner != "" {
			lines = append(lines, fmt.Sprintf("port %d: %s", p.Port, p.Banner))
		} else {
			lines = append(lines, fmt.Sprintf("port %d", p.Port))
		}
	}
	return strings.Join(lines, "\n")
}

// DemoHosts defines fake hosts with real MAC addresses so demo uses the OUI lookup
var DemoHosts = []HostResult{
	{
		IP: "192.168.1.1", Hardware: "f0:9f:c2:1a:22:01",
		Ports: []PortInfo{
			{22, "SSH-2.0-OpenSSH_8.4"},
			{80, "UniFi OS 3.2.12"},
			{443, ""},
		},
	},
	{
		IP: "192.168.1.50", Hardware: "48:b0:2d:5e:a3:10",
		Ports: []PortInfo{
			{22, "SSH-2.0-OpenSSH_8.9p1 Ubuntu-3ubuntu0.6"},
			{5432, ""},
		},
	},
	{
		IP: "192.168.1.75", Hardware: "9c:b7:0d:0a:3f:12",
		Ports: []PortInfo{
			{22, "SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.11"},
		},
	},
	{
		IP: "192.168.1.100", Hardware: "f0:ee:7a:ab:cd:ef",
		Ports: []PortInfo{
			{80, "Apache/2.4.56"},
			{443, ""},
			{8080, "Jetty 11.0.15"},
		},
	},
	{
		IP: "10.0.0.42", Hardware: "d8:3a:dd:11:22:33",
		Ports: []PortInfo{
			{22, "SSH-2.0-OpenSSH_9.2p1 Debian-2+deb12u2"},
			{80, "lighttpd/1.4.69"},
			{1883, ""},
		},
	},
}

// ScanProgress represents progress updates during scanning
type ScanProgress struct {
	Host         string // Host found (IP, hostname, port)
	Scanned      int    // Number of hosts scanned so far in the sweep phase
	Total        int    // Total hosts to scan in the sweep phase
	Phase        string // "neighbors" or "sweep"
	PhaseScanned int    // Phase-local progress
	PhaseTotal   int    // Phase-local total
}

// Scanner abstracts network scanning so real and demo modes share the same code path.
type Scanner interface {
	ScanNetwork(ifaceName, subnet string, progressChan chan<- ScanProgress)
}

// NetScanner performs real network scanning (TCP connect, ARP, banner grab).
type NetScanner struct{}

// DemoScanner simulates scanning with fake host data.
type DemoScanner struct{}

type portResult struct {
	port   int
	banner string
}

// totalScanHosts returns the number of IPv4 hosts that will actually be scanned.
// For normal subnets, excludes network+broadcast. For /31 and /32, keeps all addresses.
func totalScanHosts(ipnet *net.IPNet) int {
	ones, bits := ipnet.Mask.Size()
	hostBits := bits - ones

	// Non-IPv4 fallback keeps prior behavior.
	if bits != 32 {
		return 1 << uint(hostBits)
	}

	switch {
	case hostBits <= 0:
		return 1
	case hostBits == 1:
		return 2
	default:
		return (1 << uint(hostBits)) - 2
	}
}

// ScanNetwork scans a real subnet with controlled concurrency for smooth progress.
func (s *NetScanner) ScanNetwork(ifaceName, subnet string, progressChan chan<- ScanProgress) {
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return
	}

	totalHosts := totalScanHosts(ipnet)

	skipIPs := s.runNeighborDiscoveryPhase(ifaceName, ipnet, totalHosts, progressChan)
	s.runSubnetSweepPhase(ifaceName, ipnet, totalHosts, skipIPs, progressChan)
	close(progressChan) // Signal completion
}

// runNeighborDiscoveryPhase discovers hosts already visible in the ARP/neighbor
// table and emits them immediately, returning IPs to skip during the full
// subnet sweep.
func (s *NetScanner) runNeighborDiscoveryPhase(ifaceName string, subnet *net.IPNet, totalHosts int, progressChan chan<- ScanProgress) map[string]struct{} {
	discovered := DiscoverVisibleNeighbors(ifaceName, subnet)
	skipIPs := make(map[string]struct{}, len(discovered))
	for _, neighbor := range discovered {
		skipIPs[neighbor.IP] = struct{}{}
	}

	if len(discovered) == 0 {
		select {
		case progressChan <- ScanProgress{
			Scanned:      0,
			Total:        totalHosts,
			Phase:        "neighbors",
			PhaseScanned: 0,
			PhaseTotal:   0,
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

			hostInfo := scanHostWithKnownMAC(ifaceName, neighbor.IP, neighbor.MAC)
			if hostInfo == "" {
				hardware := VendorFromMAC(neighbor.MAC)
				hostInfo = FormatHost(HostResult{
					IP:       neighbor.IP,
					Hardware: hardware,
				})
			}

			progressMu.Lock()
			phaseDone++
			currentDone := phaseDone
			progressMu.Unlock()

			select {
			case progressChan <- ScanProgress{
				Host:         hostInfo,
				Scanned:      0,
				Total:        totalHosts,
				Phase:        "neighbors",
				PhaseScanned: currentDone,
				PhaseTotal:   len(discovered),
			}:
			default:
			}
		}(neighbor)
	}
	wg.Wait()

	return skipIPs
}

func scanOpenPorts(ip string) []portResult {
	var wg sync.WaitGroup
	var mutex sync.Mutex
	var results []portResult

	// Scan all ports in parallel for this host
	for _, port := range CommonPorts {
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
		return VendorFromMAC(knownMAC)
	}

	if targetIP == nil {
		return ""
	}

	mac := ResolveMAC(ifaceName, targetIP)
	if mac == "" {
		mac = LookupMACFromCache(targetIP.String())
	}
	if mac == "" {
		return ""
	}
	return VendorFromMAC(mac)
}

func scanHostWithKnownMAC(ifaceName string, ip string, knownMAC string) string {
	results := scanOpenPorts(ip)
	if len(results) == 0 {
		return ""
	}

	// Resolve hardware manufacturer via known MAC hint or ARP/cache fallback.
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
	result := HostResult{IP: ip, Hardware: hardware}
	for _, r := range results {
		result.Ports = append(result.Ports, PortInfo{Port: r.port, Banner: r.banner})
	}

	return FormatHost(result)
}

// runSubnetSweepPhase scans the whole subnet and skips hosts already found
// via neighbor discovery.
func (s *NetScanner) runSubnetSweepPhase(ifaceName string, subnet *net.IPNet, totalHosts int, skipIPs map[string]struct{}, progressChan chan<- ScanProgress) {
	var wg sync.WaitGroup
	var mutex sync.Mutex
	scanned := 0

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
				hostInfo = scanHost(ifaceName, currentIP)
			}

			mutex.Lock()
			scanned++
			currentScanned := scanned
			mutex.Unlock()

			// Send progress update for every host
			select {
			case progressChan <- ScanProgress{
				Host:         hostInfo,
				Scanned:      currentScanned,
				Total:        totalHosts,
				Phase:        "sweep",
				PhaseScanned: currentScanned,
				PhaseTotal:   totalHosts,
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
func scanHost(ifaceName string, ip string) string {
	return scanHostWithKnownMAC(ifaceName, ip, "")
}

// grabServiceBanner reads the actual response from a service.
// For push-banner protocols (SSH, SMTP, FTP) it just reads.
// For HTTP ports it sends a HEAD request to get the Server header.
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

// ScanNetwork simulates a scan with fake hosts for demo mode.
func (s *DemoScanner) ScanNetwork(ifaceName, subnet string, progressChan chan<- ScanProgress) {
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		close(progressChan)
		return
	}

	totalHosts := totalScanHosts(ipnet)

	// Pick which demo hosts belong to this subnet, resolve MAC via OUI
	var subnetHosts []HostResult
	for _, h := range DemoHosts {
		ip := net.ParseIP(h.IP)
		if ip != nil && ipnet.Contains(ip) {
			resolved := h
			resolved.Hardware = VendorFromMAC(h.Hardware)
			subnetHosts = append(subnetHosts, resolved)
		}
	}

	// Simulate "already visible" neighbors first (phase 1), then sweep.
	neighborCount := 0
	if len(subnetHosts) > 0 {
		neighborCount = 1
		if len(subnetHosts) > 2 {
			neighborCount = 2
		}
	}

	neighbors := subnetHosts[:neighborCount]
	remaining := subnetHosts[neighborCount:]
	for i, h := range neighbors {
		time.Sleep(180 * time.Millisecond)
		select {
		case progressChan <- ScanProgress{
			Host:         FormatHost(h),
			Scanned:      0,
			Total:        totalHosts,
			Phase:        "neighbors",
			PhaseScanned: i + 1,
			PhaseTotal:   neighborCount,
		}:
		}
	}
	if neighborCount == 0 {
		select {
		case progressChan <- ScanProgress{
			Scanned:      0,
			Total:        totalHosts,
			Phase:        "neighbors",
			PhaseScanned: 0,
			PhaseTotal:   0,
		}:
		}
	}

	// Space remaining hosts evenly across the sweep (phase 2).
	hostInterval := 0
	if len(remaining) > 0 {
		hostInterval = totalHosts / (len(remaining) + 1)
	}
	hostIdx := 0

	for i := 1; i <= totalHosts; i++ {
		time.Sleep(10 * time.Millisecond)

		host := ""
		if hostInterval > 0 && hostIdx < len(remaining) && i == hostInterval*(hostIdx+1) {
			host = FormatHost(remaining[hostIdx])
			hostIdx++
		}

		select {
		case progressChan <- ScanProgress{
			Host:         host,
			Scanned:      i,
			Total:        totalHosts,
			Phase:        "sweep",
			PhaseScanned: i,
			PhaseTotal:   totalHosts,
		}:
		}
	}

	close(progressChan)
}
