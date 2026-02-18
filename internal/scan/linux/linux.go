package linux

import (
	"os"
	"strings"

	"github.com/backendsystems/nibble/internal/scan/shared"
)

type Neighbor struct {
	IP  string
	MAC string
}

func LookupMAC(ip string) string {
	data, err := os.ReadFile("/proc/net/arp")
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 4 && fields[0] == ip && fields[3] != "00:00:00:00:00:00" {
			return shared.NormalizeMAC(fields[3])
		}
	}

	return ""
}

func Neighbors(ifaceName string) []Neighbor {
	data, err := os.ReadFile("/proc/net/arp")
	if err != nil {
		return nil
	}

	rows := make([]Neighbor, 0)
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		ip := fields[0]
		mac := shared.NormalizeMAC(fields[3])
		dev := fields[5]
		if dev != ifaceName || mac == "" {
			continue
		}

		rows = append(rows, Neighbor{IP: ip, MAC: mac})
	}

	return rows
}
