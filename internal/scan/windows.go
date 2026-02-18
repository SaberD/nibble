package scan

import (
	"os/exec"
	"regexp"
	"strings"
)

func lookupMacFromWindowsArp(ip string) string {
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
		mac := normalizeMac(match[1])
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

	// Example:
	// 192.168.1.1          00-11-22-33-44-55     dynamic
	re := regexp.MustCompile(`(?i)^\s*([0-9]{1,3}(?:\.[0-9]{1,3}){3})\s+([0-9a-f]{2}(?:-[0-9a-f]{2}){5})\s+`)

	var rows []NeighborEntry
	for _, line := range strings.Split(string(out), "\n") {
		match := re.FindStringSubmatch(strings.TrimSpace(line))
		if len(match) != 3 {
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
