package demo

import (
	"net"
	"time"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scan"
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
	hostOnly := len(s.Ports) == 0
	hosts := hostsForInterface(ifaceName)
	neighborDelay, sweepDelay := demoDelaysForInterface(ifaceName)

	// Pick which demo hosts belong to this subnet.
	var subnetHosts []scanner.HostResult
	for _, h := range hosts {
		ip := net.ParseIP(h.IP)
		if ip == nil || !ipnet.Contains(ip) {
			continue
		}
		resolved := scanner.HostResult{
			IP:       h.IP,
			Hardware: scan.VendorFromMac(h.Hardware),
		}
		if !hostOnly {
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
		time.Sleep(neighborDelay)
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
		time.Sleep(sweepDelay)

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

func hostsForInterface(ifaceName string) []Host {
	if ifaceName == "wlan0" {
		return WiFiHosts
	}
	return Hosts
}

func demoDelaysForInterface(ifaceName string) (neighborDelay, sweepDelay time.Duration) {
	if ifaceName == "wlan0" {
		return 260 * time.Millisecond, 20 * time.Millisecond
	}
	return 180 * time.Millisecond, 10 * time.Millisecond
}

func selectedPorts(configured []int) []int {
	if configured != nil {
		return configured
	}
	return ports.DefaultPorts()
}
