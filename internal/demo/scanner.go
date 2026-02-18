package demo

import (
	"net"
	"time"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scanner"
)

// DemoScanner simulates a scan with fake host data.
type DemoScanner struct {
	Ports []int
}

func (s *DemoScanner) ScanNetwork(ifaceName, subnet string, progressChan chan<- scanner.ProgressUpdate) {
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		close(progressChan)
		return
	}

	totalHosts := scanner.TotalScanHosts(ipnet)
	selected := selectedPorts(s.Ports)
	selectedSet := make(map[int]struct{}, len(selected))
	for _, p := range selected {
		selectedSet[p] = struct{}{}
	}

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
			if _, ok := selectedSet[p.Port]; !ok {
				continue
			}
			resolved.Ports = append(resolved.Ports, scanner.PortInfo{
				Port:   p.Port,
				Banner: p.Banner,
			})
		}
		if len(resolved.Ports) == 0 {
			continue
		}
		subnetHosts = append(subnetHosts, resolved)
	}

	// Emit nearby hosts first, then run the full sweep.
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
		progressChan <- scanner.NeighborProgress{
			Host:       scanner.FormatHost(h),
			TotalHosts: totalHosts,
			Seen:       i + 1,
			Total:      neighborCount,
		}
	}
	if neighborCount == 0 {
		progressChan <- scanner.NeighborProgress{
			TotalHosts: totalHosts,
			Seen:       0,
			Total:      0,
		}
	}

	// Spread remaining hosts across the sweep.
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

		progressChan <- scanner.SweepProgress{
			Host:       host,
			TotalHosts: totalHosts,
			Scanned:    i,
			Total:      totalHosts,
		}
	}

	close(progressChan)
}

func selectedPorts(configured []int) []int {
	if len(configured) > 0 {
		return configured
	}
	return ports.DefaultPorts()
}
