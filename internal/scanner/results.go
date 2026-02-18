package scanner

import (
	"fmt"
	"strings"
)

// PortInfo holds a port number and its service banner.
type PortInfo struct {
	Port   int
	Banner string
}

// HostResult holds all scan info for a single host.
type HostResult struct {
	IP       string
	Hardware string
	Ports    []PortInfo
}

// FormatHost renders a HostResult into the display string.
func FormatHost(h HostResult) string {
	var lines []string
	if h.Hardware != "" {
		lines = append(lines, fmt.Sprintf("%s - %s", h.IP, h.Hardware))
	} else {
		lines = append(lines, h.IP)
	}
	for _, p := range h.Ports {
		if p.Banner != "" {
			lines = append(lines, fmt.Sprintf("port %d: %s", p.Port, p.Banner))
		} else {
			lines = append(lines, fmt.Sprintf("port %d", p.Port))
		}
	}
	return strings.Join(lines, "\n")
}
