package scan

import (
	"net"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scanner"
)

// NetScanner performs real network scanning (TCP connect, ARP, banner grab)
type NetScanner struct {
	Ports []int
}

// ScanNetwork scans a real subnet with controlled concurrency for smooth progress
func (s *NetScanner) ScanNetwork(ifaceName, subnet string, progressChan chan<- scanner.ProgressUpdate) {
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return
	}

	totalHosts := scanner.TotalScanHosts(ipnet)
	skipIPs := s.neighborDiscovery(ifaceName, ipnet, totalHosts, progressChan)
	s.subnetSweep(ifaceName, ipnet, totalHosts, skipIPs, progressChan)

	close(progressChan)
}

func (s *NetScanner) ports() (out []int) {
	out = ports.DefaultPorts()
	if s.Ports != nil {
		out = s.Ports
	}
	return out
}
