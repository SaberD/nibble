package scan

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/backendsystems/nibble/internal/scanner"
)

const (
	neighborPhaseMaxWorkers = 16
	sweepPhaseMaxWorkers    = 100
)

// neighborDiscovery emits hosts already visible in neighbor tables
// and returns IPs that should be skipped in the full sweep
func (s *NetScanner) neighborDiscovery(ifaceName string, subnet *net.IPNet, totalHosts int, progressChan chan<- scanner.ProgressUpdate) map[string]struct{} {
	neighbors := visibleNeighbors(ifaceName, subnet)
	skipIPs := buildSkipMap(neighbors)
	if len(neighbors) == 0 {
		emitNeighborProgress(progressChan, scanner.NeighborProgress{TotalHosts: totalHosts})
		return skipIPs
	}

	workerCount := neighborPhaseMaxWorkers
	if len(neighbors) < workerCount {
		workerCount = len(neighbors)
	}

	ports := s.ports()
	jobs := make(chan NeighborEntry)
	var wg sync.WaitGroup
	var seenCount atomic.Int64

	for range workerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for neighbor := range jobs {
				processNeighborJob(ifaceName, neighbor, ports, totalHosts, len(neighbors), &seenCount, progressChan)
			}
		}()
	}

	for _, neighbor := range neighbors {
		jobs <- neighbor
	}
	close(jobs)

	wg.Wait()
	return skipIPs
}

// subnetSweep scans the subnet and skips hosts found in neighbor discovery
func (s *NetScanner) subnetSweep(ifaceName string, subnet *net.IPNet, totalHosts int, skipIPs map[string]struct{}, progressChan chan<- scanner.ProgressUpdate) {
	ports := s.ports()
	jobs := make(chan string, sweepPhaseMaxWorkers)
	var wg sync.WaitGroup
	var scanned atomic.Int64

	for range sweepPhaseMaxWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for currentIP := range jobs {
				processSweepJob(ifaceName, currentIP, ports, skipIPs, totalHosts, &scanned, progressChan)
			}
		}()
	}

	for ip := subnet.IP.Mask(subnet.Mask); subnet.Contains(ip); incrementIP(ip) {
		if skipIp4(ip, subnet) {
			continue
		}

		jobs <- ip.String()
	}
	close(jobs)

	wg.Wait()
}

func processNeighborJob(ifaceName string, neighbor NeighborEntry, ports []int, totalHosts, totalNeighbors int, seenCount *atomic.Int64, progressChan chan<- scanner.ProgressUpdate) {
	hostInfo := scanHostMac(ifaceName, neighbor.IP, neighbor.MAC, ports)
	if hostInfo == "" {
		hostInfo = formatHost(neighbor)
	}

	currentSeen := int(seenCount.Add(1))

	emitNeighborProgress(progressChan, scanner.NeighborProgress{
		Host:       hostInfo,
		TotalHosts: totalHosts,
		Seen:       currentSeen,
		Total:      totalNeighbors,
	})
}

func processSweepJob(ifaceName, currentIP string, ports []int, skipIPs map[string]struct{}, totalHosts int, scanned *atomic.Int64, progressChan chan<- scanner.ProgressUpdate) {
	hostInfo := ""
	if _, alreadyFound := skipIPs[currentIP]; !alreadyFound {
		hostInfo = scanHost(ifaceName, currentIP, ports)
	}

	currentScanned := int(scanned.Add(1))

	progressChan <- scanner.SweepProgress{
		Host:       hostInfo,
		TotalHosts: totalHosts,
		Scanned:    currentScanned,
		Total:      totalHosts,
	}
}

func buildSkipMap(neighbors []NeighborEntry) map[string]struct{} {
	skipIPs := make(map[string]struct{}, len(neighbors))
	for _, neighbor := range neighbors {
		skipIPs[neighbor.IP] = struct{}{}
	}
	return skipIPs
}

func formatHost(neighbor NeighborEntry) string {
	return scanner.FormatHost(scanner.HostResult{
		IP:       neighbor.IP,
		Hardware: VendorFromMac(neighbor.MAC),
	})
}

func emitNeighborProgress(progressChan chan<- scanner.ProgressUpdate, progress scanner.NeighborProgress) {
	select {
	case progressChan <- progress:
	default:
	}
}

func incrementIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] > 0 {
			return
		}
	}
}

func skipIp4(ip net.IP, subnet *net.IPNet) bool {
	ip4 := ip.To4()
	base := subnet.IP.To4()
	mask := subnet.Mask
	if ip4 == nil || base == nil || len(mask) != net.IPv4len {
		return false
	}
	// loop
	if ip4.Equal(base.Mask(mask)) {
		return true
	}

	broadcast := net.IPv4(
		base[0]|^mask[0],
		base[1]|^mask[1],
		base[2]|^mask[2],
		base[3]|^mask[3],
	)
	// broadcast
	return ip4.Equal(broadcast)
}
