package demo

import (
	"net"
	"time"

	"github.com/backendsystems/nibble/internal/scanner"
)

// DemoScanner simulates a scan with fake host data.
type DemoScanner struct{}

func (s *DemoScanner) ScanNetwork(ifaceName, subnet string, progressChan chan<- scanner.ProgressUpdate) {
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		close(progressChan)
		return
	}

	totalHosts := scanner.TotalScanHosts(ipnet)

	// Pick which demo hosts belong to this subnet.
	var subnetHosts []scanner.HostResult
	for _, h := range Hosts {
		ip := net.ParseIP(h.IP)
		if ip == nil || !ipnet.Contains(ip) {
			continue
		}
		resolved := scanner.HostResult{
			IP:       h.IP,
			Hardware: h.Hardware,
		}
		for _, p := range h.Ports {
			resolved.Ports = append(resolved.Ports, scanner.PortInfo{
				Port:   p.Port,
				Banner: p.Banner,
			})
		}
		subnetHosts = append(subnetHosts, resolved)
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
		case progressChan <- scanner.NeighborProgress{
			Host:       scanner.FormatHost(h),
			TotalHosts: totalHosts,
			Seen:       i + 1,
			Total:      neighborCount,
		}:
		}
	}
	if neighborCount == 0 {
		select {
		case progressChan <- scanner.NeighborProgress{
			TotalHosts: totalHosts,
			Seen:       0,
			Total:      0,
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
			host = scanner.FormatHost(remaining[hostIdx])
			hostIdx++
		}

		select {
		case progressChan <- scanner.SweepProgress{
			Host:       host,
			TotalHosts: totalHosts,
			Scanned:    i,
			Total:      totalHosts,
		}:
		}
	}

	close(progressChan)
}
