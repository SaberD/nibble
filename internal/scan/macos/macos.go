package macos

import (
	"net/netip"
	"os/exec"
	"strings"

	"github.com/backendsystems/nibble/internal/scan/shared"
)

type Neighbor struct {
	IP    string
	MAC   string
	Iface string
}

func LookupMAC(ip string) string {
	out, err := exec.Command("arp", "-an").Output()
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(out), "\n") {
		row, ok := parseRow(line)
		if !ok || row.IP != ip {
			continue
		}
		if row.MAC != "00:00:00:00:00:00" {
			return row.MAC
		}
	}

	return ""
}

func Neighbors(ifaceName string) []Neighbor {
	out, err := exec.Command("arp", "-an").Output()
	if err != nil {
		return nil
	}

	rows := make([]Neighbor, 0)
	for _, line := range strings.Split(string(out), "\n") {
		row, ok := parseRow(line)
		if !ok || row.Iface != ifaceName {
			continue
		}
		rows = append(rows, row)
	}

	return rows
}

func parseRow(line string) (Neighbor, bool) {
	fields := strings.Fields(line)
	if len(fields) < 6 {
		return Neighbor{}, false
	}
	if fields[2] != "at" || fields[4] != "on" {
		return Neighbor{}, false
	}

	ip := strings.Trim(fields[1], "()")
	addr, err := netip.ParseAddr(ip)
	if err != nil || !addr.Is4() {
		return Neighbor{}, false
	}

	mac := shared.NormalizeMAC(fields[3])
	if mac == "" {
		return Neighbor{}, false
	}

	return Neighbor{
		IP:    addr.String(),
		MAC:   mac,
		Iface: fields[5],
	}, true
}
