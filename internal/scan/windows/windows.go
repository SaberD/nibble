package windows

import (
	"net/netip"
	"os/exec"
	"strings"

	"github.com/backendsystems/nibble/internal/scan/shared"
)

type Neighbor struct {
	IP  string
	MAC string
}

func LookupMAC(ip string) string {
	out, err := exec.Command("arp", "-a").Output()
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != ip {
			continue
		}
		mac := shared.NormalizeMAC(fields[1])
		if mac != "" && mac != "00:00:00:00:00:00" {
			return mac
		}
	}

	return ""
}

func Neighbors() []Neighbor {
	out, err := exec.Command("arp", "-a").Output()
	if err != nil {
		return nil
	}

	rows := make([]Neighbor, 0)
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		ip, err := netip.ParseAddr(fields[0])
		if err != nil || !ip.Is4() {
			continue
		}
		mac := shared.NormalizeMAC(fields[1])
		if mac == "" {
			continue
		}
		rows = append(rows, Neighbor{IP: ip.String(), MAC: mac})
	}

	return rows
}
