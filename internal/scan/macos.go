package scan

import (
	"net/netip"
	"os/exec"
	"strings"
)

func lookupMacFromDarwinArp(ip string) string {
	out, err := exec.Command("arp", "-an").Output()
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(out), "\n") {
		row, ok := parseDarwinArpLine(line)
		if !ok || row.IP != ip {
			continue
		}
		if row.MAC != "00:00:00:00:00:00" {
			return row.MAC
		}
	}
	return ""
}

func readNeighborsDarwinArp(ifaceName string) []NeighborEntry {
	out, err := exec.Command("arp", "-an").Output()
	if err != nil {
		return nil
	}

	var rows []NeighborEntry
	for _, line := range strings.Split(string(out), "\n") {
		row, ok := parseDarwinArpLine(line)
		if !ok || row.Iface != ifaceName {
			continue
		}
		rows = append(rows, NeighborEntry{IP: row.IP, MAC: row.MAC})
	}
	return rows
}

type darwinArpRow struct {
	IP    string
	MAC   string
	Iface string
}

func parseDarwinArpLine(line string) (darwinArpRow, bool) {
	fields := strings.Fields(line)
	if len(fields) < 6 {
		return darwinArpRow{}, false
	}
	if fields[2] != "at" || fields[4] != "on" {
		return darwinArpRow{}, false
	}

	ip := strings.Trim(fields[1], "()")
	addr, err := netip.ParseAddr(ip)
	if err != nil || !addr.Is4() {
		return darwinArpRow{}, false
	}

	mac := normalizeMac(fields[3])
	if mac == "" {
		return darwinArpRow{}, false
	}

	return darwinArpRow{
		IP:    addr.String(),
		MAC:   mac,
		Iface: fields[5],
	}, true
}
