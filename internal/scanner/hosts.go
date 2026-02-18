package scanner

import (
	"net"
	"net/netip"
	"strings"
)

// TotalScanHosts returns the number of IPv4 hosts that will actually be scanned.
func TotalScanHosts(ipnet *net.IPNet) int {
	ones, bits := ipnet.Mask.Size()
	hostBits := bits - ones

	// non-ip4 fallback keeps prior behavior.
	if bits != 32 {
		return 1 << uint(hostBits)
	}

	// /32 has one host.
	if hostBits <= 0 {
		return 1
	}

	totalHosts := 1 << uint(hostBits)

	// /31 keeps both addresses as usable hosts.
	if hostBits == 1 {
		return totalHosts
	}

	// Larger subnets skip network and broadcast addresses.
	return totalHosts - 2
}

// FirstIp4 returns the first IPv4 CIDR string from interface addresses.
func FirstIp4(addrs []net.Addr) string {
	for _, addr := range addrs {
		addrText := addr.String()
		ip, ok := parseAddr(addrText)
		if !ok || !ip.Is4() {
			continue
		}

		// Normalize plain IPv4 values for ParseCIDR callers.
		if strings.Contains(addrText, "/") {
			return addrText
		}
		return ip.String() + "/32"
	}
	return ""
}

func parseAddr(s string) (netip.Addr, bool) {
	// addresses may be CIDR "192.168.1.10/24"
	prefix, err := netip.ParsePrefix(s)
	if err == nil {
		return prefix.Addr(), true
	}

	// or plain "192.168.1.10"
	ip, err := netip.ParseAddr(s)
	if err == nil {
		return ip, true
	}

	return netip.Addr{}, false
}
