package scan

import (
	"encoding/csv"
	"net"
	"net/netip"
	"os"
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

// LookupMACFromCache reads /proc/net/arp to find a MAC without needing root.
// The kernel populates this cache after any successful TCP connection.
func LookupMACFromCache(ip string) string {
	data, err := os.ReadFile("/proc/net/arp")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 4 && fields[0] == ip && fields[3] != "00:00:00:00:00:00" {
			return fields[3]
		}
	}
	return ""
}
