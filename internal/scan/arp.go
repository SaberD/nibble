package scan

import (
	"net"
	"net/netip"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/mdlayher/arp"
)

// resolveMac does an ARP resolve for a single IP and returns the MAC string.
func resolveMac(ifaceName string, targetIP net.IP) string {
	netIface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return ""
	}

	client, err := arp.Dial(netIface)
	if err != nil {
		return ""
	}
	defer client.Close()

	client.SetDeadline(time.Now().Add(200 * time.Millisecond))

	addr, ok := netip.AddrFromSlice(targetIP.To4())
	if !ok {
		return ""
	}

	mac, err := client.Resolve(addr)
	if err != nil {
		return ""
	}

	return mac.String()
}

// lookupMacFromCache reads the OS ARP cache to find a MAC without needing root.
// Linux uses /proc/net/arp; Windows uses `arp -a`; macOS uses `arp -an`.
func lookupMacFromCache(ip string) string {
	if runtime.GOOS == "windows" {
		return lookupMacFromWindowsArp(ip)
	}
	if runtime.GOOS == "darwin" {
		return lookupMacFromDarwinArp(ip)
	}

	data, err := os.ReadFile("/proc/net/arp")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 4 && fields[0] == ip && fields[3] != "00:00:00:00:00:00" {
			return normalizeMac(fields[3])
		}
	}
	return ""
}

func normalizeMac(mac string) string {
	mac = strings.ToLower(strings.TrimSpace(strings.ReplaceAll(mac, "-", ":")))
	parts := strings.Split(mac, ":")
	if len(parts) != 6 {
		return ""
	}

	for i, part := range parts {
		if len(part) == 0 || len(part) > 2 {
			return ""
		}
		v, err := strconv.ParseUint(part, 16, 8)
		if err != nil {
			return ""
		}
		parts[i] = strings.ToLower(strconv.FormatUint(v, 16))
		if len(parts[i]) == 1 {
			parts[i] = "0" + parts[i]
		}
	}
	return strings.Join(parts, ":")
}

// NeighborEntry is a visible L2/L3 neighbor from the host ARP/neighbor table.
type NeighborEntry struct {
	IP  string
	MAC string
}

// visibleNeighbors returns neighbors currently visible in the OS ARP
// table for the selected interface and subnet.
func visibleNeighbors(ifaceName string, subnet *net.IPNet) []NeighborEntry {
	var rows []NeighborEntry
	switch runtime.GOOS {
	case "windows":
		rows = readNeighborsWindowsArp()
	case "darwin":
		rows = readNeighborsDarwinArp(ifaceName)
	default:
		rows = readNeighborsProcArp(ifaceName)
	}

	seen := make(map[string]struct{})
	var out []NeighborEntry
	for _, row := range rows {
		ip := net.ParseIP(row.IP)
		if ip == nil || ip.To4() == nil || !subnet.Contains(ip) {
			continue
		}
		if row.MAC == "" || row.MAC == "00:00:00:00:00:00" {
			continue
		}
		if strings.EqualFold(row.MAC, "ff:ff:ff:ff:ff:ff") {
			continue
		}
		if isSubnetBroadcastIpv4(ip, subnet) {
			continue
		}
		if _, ok := seen[row.IP]; ok {
			continue
		}
		seen[row.IP] = struct{}{}
		out = append(out, row)
	}
	return out
}

func isSubnetBroadcastIpv4(ip net.IP, subnet *net.IPNet) bool {
	ip4 := ip.To4()
	base := subnet.IP.To4()
	mask := subnet.Mask
	if ip4 == nil || base == nil || len(mask) != net.IPv4len {
		return false
	}
	broadcast := net.IPv4(
		base[0]|^mask[0],
		base[1]|^mask[1],
		base[2]|^mask[2],
		base[3]|^mask[3],
	)
	return ip4.Equal(broadcast)
}

func readNeighborsProcArp(ifaceName string) []NeighborEntry {
	data, err := os.ReadFile("/proc/net/arp")
	if err != nil {
		return nil
	}

	var out []NeighborEntry
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		ip := fields[0]
		mac := normalizeMac(fields[3])
		dev := fields[5]
		if dev != ifaceName || mac == "" {
			continue
		}
		out = append(out, NeighborEntry{IP: ip, MAC: mac})
	}
	return out
}
