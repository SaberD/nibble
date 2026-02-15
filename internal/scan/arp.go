package scan

import (
	"encoding/csv"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"strconv"
	"runtime"
	"regexp"
	"strings"
	"time"

	_ "embed"

	"github.com/mdlayher/arp"
)

//go:embed oui.csv
var ouiCSV string

// OUI maps MAC prefixes (lowercase, colon-separated) to manufacturer names.
// Loaded from the embedded IEEE OUI database at init time.
var OUI map[string]string

func init() {
	OUI = make(map[string]string, 40000)
	r := csv.NewReader(strings.NewReader(ouiCSV))
	records, err := r.ReadAll()
	if err != nil {
		return
	}
	for _, rec := range records[1:] { // skip header
		if len(rec) < 3 {
			continue
		}
		// Assignment is 6 hex chars like "286FB9", convert to "28:6f:b9"
		hex := strings.ToLower(rec[1])
		if len(hex) != 6 {
			continue
		}
		prefix := hex[0:2] + ":" + hex[2:4] + ":" + hex[4:6]
		OUI[prefix] = rec[2]
	}
}

// ResolveMAC does an ARP resolve for a single IP and returns the MAC string.
func ResolveMAC(ifaceName string, targetIP net.IP) string {
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

// VendorFromMAC returns the hardware manufacturer from a MAC address OUI prefix
func VendorFromMAC(mac string) string {
	mac = strings.ToLower(mac)
	if len(mac) >= 8 {
		if vendor, ok := OUI[mac[:8]]; ok {
			return vendor
		}
	}
	if mac != "" {
		return strings.ToUpper(mac)
	}
	return ""
}

// LookupMACFromCache reads the OS ARP cache to find a MAC without needing root.
// Linux uses /proc/net/arp; Windows uses `arp -a`; macOS uses `arp -an`.
func LookupMACFromCache(ip string) string {
	if runtime.GOOS == "windows" {
		return lookupMACFromWindowsARP(ip)
	}
	if runtime.GOOS == "darwin" {
		return lookupMACFromDarwinARP(ip)
	}

	data, err := os.ReadFile("/proc/net/arp")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 4 && fields[0] == ip && fields[3] != "00:00:00:00:00:00" {
			return normalizeMAC(fields[3])
		}
	}
	return ""
}

func lookupMACFromWindowsARP(ip string) string {
	out, err := exec.Command("arp", "-a").Output()
	if err != nil {
		return ""
	}

	// Windows arp -a shows entries like:
	//   192.168.1.1           00-11-22-33-44-55     dynamic
	re := regexp.MustCompile(`(?i)^` + regexp.QuoteMeta(ip) + `\s+([0-9a-f]{2}(?:-[0-9a-f]{2}){5})\s+`)
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		match := re.FindStringSubmatch(line)
		if len(match) != 2 {
			continue
		}
		mac := normalizeMAC(match[1])
		if mac != "" && mac != "00:00:00:00:00:00" {
			return mac
		}
	}
	return ""
}

func lookupMACFromDarwinARP(ip string) string {
	out, err := exec.Command("arp", "-an").Output()
	if err != nil {
		return ""
	}

	// macOS arp -an shows entries like:
	//   ? (192.168.1.1) at 0:11:22:33:44:55 on en0 ifscope [ethernet]
	re := regexp.MustCompile(`(?i)\(` + regexp.QuoteMeta(ip) + `\)\s+at\s+([0-9a-f:]{11,17})\s+on\s+`)
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		match := re.FindStringSubmatch(line)
		if len(match) != 2 {
			continue
		}
		mac := normalizeMAC(match[1])
		if mac != "" && mac != "00:00:00:00:00:00" {
			return mac
		}
	}
	return ""
}

func normalizeMAC(mac string) string {
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

// DiscoverVisibleNeighbors returns neighbors currently visible in the OS ARP
// table for the selected interface and subnet.
func DiscoverVisibleNeighbors(ifaceName string, subnet *net.IPNet) []NeighborEntry {
	var rows []NeighborEntry
	switch runtime.GOOS {
	case "windows":
		rows = discoverNeighborsFromWindowsARP()
	case "darwin":
		rows = discoverNeighborsFromDarwinARP(ifaceName)
	default:
		rows = discoverNeighborsFromProcARP(ifaceName)
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
		if isSubnetBroadcastIPv4(ip, subnet) {
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

func isSubnetBroadcastIPv4(ip net.IP, subnet *net.IPNet) bool {
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

func discoverNeighborsFromProcARP(ifaceName string) []NeighborEntry {
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
		mac := normalizeMAC(fields[3])
		dev := fields[5]
		if dev != ifaceName || mac == "" {
			continue
		}
		out = append(out, NeighborEntry{IP: ip, MAC: mac})
	}
	return out
}

func discoverNeighborsFromWindowsARP() []NeighborEntry {
	out, err := exec.Command("arp", "-a").Output()
	if err != nil {
		return nil
	}

	// Example:
	// 192.168.1.1          00-11-22-33-44-55     dynamic
	re := regexp.MustCompile(`(?i)^\s*([0-9]{1,3}(?:\.[0-9]{1,3}){3})\s+([0-9a-f]{2}(?:-[0-9a-f]{2}){5})\s+`)

	var rows []NeighborEntry
	for _, line := range strings.Split(string(out), "\n") {
		match := re.FindStringSubmatch(strings.TrimSpace(line))
		if len(match) != 3 {
			continue
		}
		mac := normalizeMAC(match[2])
		if mac == "" {
			continue
		}
		rows = append(rows, NeighborEntry{IP: match[1], MAC: mac})
	}
	return rows
}

func discoverNeighborsFromDarwinARP(ifaceName string) []NeighborEntry {
	out, err := exec.Command("arp", "-an").Output()
	if err != nil {
		return nil
	}

	// Example:
	// ? (192.168.1.1) at 0:11:22:33:44:55 on en0 ifscope [ethernet]
	re := regexp.MustCompile(`(?i)\(([0-9]{1,3}(?:\.[0-9]{1,3}){3})\)\s+at\s+([0-9a-f:]{11,17})\s+on\s+(\S+)`)

	var rows []NeighborEntry
	for _, line := range strings.Split(string(out), "\n") {
		match := re.FindStringSubmatch(strings.TrimSpace(line))
		if len(match) != 4 {
			continue
		}
		if match[3] != ifaceName {
			continue
		}
		mac := normalizeMAC(match[2])
		if mac == "" {
			continue
		}
		rows = append(rows, NeighborEntry{IP: match[1], MAC: mac})
	}
	return rows
}
