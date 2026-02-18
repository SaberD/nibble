package scan

import (
	"os/exec"
	"regexp"
	"strings"
)

func lookupMacFromDarwinArp(ip string) string {
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
		mac := normalizeMac(match[1])
		if mac != "" && mac != "00:00:00:00:00:00" {
			return mac
		}
	}
	return ""
}

func readNeighborsDarwinArp(ifaceName string) []NeighborEntry {
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
		mac := normalizeMac(match[2])
		if mac == "" {
			continue
		}
		rows = append(rows, NeighborEntry{IP: match[1], MAC: mac})
	}
	return rows
}
