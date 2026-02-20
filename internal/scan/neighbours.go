package scan

import (
	"fmt"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/backendsystems/nibble/internal/scanner"
)

const portDialTimeout = 70 * time.Millisecond

type portResult struct {
	port   int
	banner string
}

func scanHost(ifaceName, ip string, ports []int) string {
	return scanHostMac(ifaceName, ip, "", ports)
}

func scanHostMac(ifaceName, ip, knownMAC string, ports []int) string {
	if len(ports) == 0 {
		// Host-only mode: ARP to check liveness (requires CAP_NET_RAW).
		// For neighbors knownMAC is already set so no ARP request is made.
		hardware := resolveHardware(ifaceName, net.ParseIP(ip), knownMAC)
		if knownMAC == "" && hardware == "" {
			return ""
		}
		return scanner.FormatHost(scanner.HostResult{IP: ip, Hardware: hardware})
	}

	results := scanOpenPorts(ip, ports)
	if len(results) == 0 {
		return ""
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].port < results[j].port
	})

	host := scanner.HostResult{
		IP:       ip,
		Hardware: resolveHardware(ifaceName, net.ParseIP(ip), knownMAC),
		Ports:    make([]scanner.PortInfo, 0, len(results)),
	}

	for _, result := range results {
		host.Ports = append(host.Ports, scanner.PortInfo{Port: result.port, Banner: result.banner})
	}

	return scanner.FormatHost(host)
}

func scanOpenPorts(ip string, ports []int) []portResult {
	var wg sync.WaitGroup
	var resultMu sync.Mutex
	results := make([]portResult, 0, len(ports))

	for _, port := range ports {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()

			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), portDialTimeout)
			if err != nil {
				return
			}
			defer conn.Close()

			resultMu.Lock()
			results = append(results, portResult{port: port, banner: getServiceBanner(conn)})
			resultMu.Unlock()
		}(port)
	}

	wg.Wait()
	return results
}

func resolveHardware(_ string, targetIP net.IP, knownMAC string) string {
	if knownMAC != "" {
		return VendorFromMac(knownMAC)
	}
	if targetIP == nil {
		return ""
	}

	mac := lookupMacFromCache(targetIP.String())
	if mac == "" {
		return ""
	}
	return VendorFromMac(mac)
}
