package scan

import (
	"net/netip"
	"os/exec"
	"strings"
)

func lookupMacFromWindowsArp(ip string) string {
	out, err := exec.Command("arp", "-a").Output()
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != ip {
			continue
		}
		mac := normalizeMac(fields[1])
		if mac != "" && mac != "00:00:00:00:00:00" {
			return mac
		}
	}
	return ""
}

func readNeighborsWindowsArp() []NeighborEntry {
	out, err := exec.Command("arp", "-a").Output()
	if err != nil {
		return nil
	}

	var rows []NeighborEntry
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		ip, err := netip.ParseAddr(fields[0])
		if err != nil || !ip.Is4() {
			continue
		}
		mac := normalizeMac(fields[1])
		if mac == "" {
			continue
		}
		rows = append(rows, NeighborEntry{IP: ip.String(), MAC: mac})
	}
	return rows
}
