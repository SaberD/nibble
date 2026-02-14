package scan

import (
	"fmt"
	"net"
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
	Host    string // Host found (IP, hostname, port)
	Scanned int    // Number of IPs scanned so far
	Total   int    // Total IPs to scan
}

// Scanner abstracts network scanning so real and demo modes share the same code path.
type Scanner interface {
	ScanNetwork(ifaceName, subnet string, progressChan chan<- ScanProgress)
}

// NetScanner performs real network scanning (TCP connect, ARP, banner grab).
type NetScanner struct{}

// DemoScanner simulates scanning with fake host data.
type DemoScanner struct{}

// ScanNetwork scans a real subnet with controlled concurrency for smooth progress.
func (s *NetScanner) ScanNetwork(ifaceName, subnet string, progressChan chan<- ScanProgress) {
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return
	}

	// Calculate total hosts
	ones, bits := ipnet.Mask.Size()
	totalHosts := 1 << uint(bits-ones)

	skipIPs := s.runMulticastDiscoveryPhase(ifaceName, ipnet, totalHosts, progressChan)
	s.runSubnetSweepPhase(ifaceName, ipnet, totalHosts, skipIPs, progressChan)
	close(progressChan) // Signal completion
}

// runMulticastDiscoveryPhase discovers hosts advertised over mDNS and emits
// them immediately, returning IPs to skip during the full subnet sweep.
func (s *NetScanner) runMulticastDiscoveryPhase(ifaceName string, subnet *net.IPNet, totalHosts int, progressChan chan<- ScanProgress) map[string]struct{} {
	discovered := discoverSSHViaMDNS(ifaceName, subnet)
	skipIPs := make(map[string]struct{}, len(discovered))
	for _, svc := range discovered {
		skipIPs[svc.IP] = struct{}{}

		hostInfo := scanHost(ifaceName, svc.IP)
		if hostInfo == "" {
			banner := "mDNS _ssh._tcp.local"
			if svc.Hostname != "" {
				banner = "mDNS " + svc.Hostname
			}
			port := int(svc.Port)
			if port == 0 {
				port = 22
			}
			hostInfo = FormatHost(HostResult{
				IP:    svc.IP,
				Ports: []PortInfo{{Port: port, Banner: banner}},
			})
		}

		select {
		case progressChan <- ScanProgress{
			Host:    hostInfo,
			Scanned: 0,
			Total:   totalHosts,
		}:
		default:
		}
	}

	return skipIPs
}

// runSubnetSweepPhase scans the whole subnet and skips hosts already found
// via multicast discovery.
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
				Host:    hostInfo,
				Scanned: currentScanned,
				Total:   totalHosts,
			}:
			}
		}(ip.String())
	}

	wg.Wait()
}

// scanHost resolves hardware via ARP, scans ports, grabs banners.
func scanHost(ifaceName string, ip string) string {
	type portResult struct {
		port   int
		banner string
	}

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

	if len(results) == 0 {
		return ""
	}

	// Resolve hardware manufacturer via ARP
	hardware := ""
	targetIP := net.ParseIP(ip)
	if targetIP != nil {
		mac := ResolveMAC(ifaceName, targetIP)
		if mac == "" {
			mac = LookupMACFromCache(ip)
		}
		if mac != "" {
			hardware = VendorFromMAC(mac)
		}
	}

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
		response := strings.TrimSpace(string(buf[:n]))
		// Clean up newlines for display
		response = strings.ReplaceAll(response, "\r\n", " ")
		response = strings.ReplaceAll(response, "\n", " ")
		if len(response) > 80 {
			response = response[:80]
		}
		return response
	}
}

// parseHTTPServer extracts the Server header from an HTTP response
func parseHTTPServer(response string) string {
	for _, line := range strings.Split(response, "\r\n") {
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "server:") {
			return strings.TrimSpace(line[7:])
		}
	}
	// No Server header, try to return the status line
	if idx := strings.Index(response, "\r\n"); idx > 0 {
		return response[:idx]
	}
	return ""
}

// ScanNetwork simulates a scan with fake hosts for demo mode.
func (s *DemoScanner) ScanNetwork(ifaceName, subnet string, progressChan chan<- ScanProgress) {
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		close(progressChan)
		return
	}

	// Calculate total hosts
	ones, bits := ipnet.Mask.Size()
	totalHosts := 1 << uint(bits-ones)

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

	// Space hosts evenly across the scan
	hostInterval := totalHosts / (len(subnetHosts) + 1)
	hostIdx := 0

	for i := 1; i <= totalHosts; i++ {
		time.Sleep(5 * time.Millisecond)

		host := ""
		if hostIdx < len(subnetHosts) && i == hostInterval*(hostIdx+1) {
			host = FormatHost(subnetHosts[hostIdx])
			hostIdx++
		}

		select {
		case progressChan <- ScanProgress{
			Host:    host,
			Scanned: i,
			Total:   totalHosts,
		}:
		}
	}

	close(progressChan)
}
