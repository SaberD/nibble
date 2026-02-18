package mainview

import (
	"fmt"
	"net"
	"strings"
)

func interfaceIcon(name string) string {
	lower := strings.ToLower(name)

	if strings.HasPrefix(lower, "docker") ||
		strings.HasPrefix(lower, "br-") ||
		strings.HasPrefix(lower, "veth") ||
		strings.HasPrefix(lower, "cni") ||
		strings.HasPrefix(lower, "flannel") ||
		strings.HasPrefix(lower, "cali") ||
		strings.HasPrefix(lower, "virbr") ||
		strings.HasPrefix(lower, "lxc") ||
		strings.HasPrefix(lower, "podman") {
		return "📦"
	}

	if strings.HasPrefix(lower, "tun") ||
		strings.HasPrefix(lower, "tap") ||
		strings.HasPrefix(lower, "utun") ||
		strings.HasPrefix(lower, "wg") ||
		strings.HasPrefix(lower, "tailscale") ||
		strings.Contains(lower, "vpn") {
		return "🔒"
	}

	if strings.HasPrefix(lower, "wl") ||
		strings.HasPrefix(lower, "wlan") ||
		strings.Contains(lower, "wi-fi") ||
		strings.Contains(lower, "wifi") ||
		strings.Contains(lower, "wireless") {
		return "📶"
	}

	if strings.HasPrefix(lower, "en") ||
		strings.HasPrefix(lower, "eth") ||
		strings.Contains(lower, "ethernet") {
		return "🔌"
	}

	return "🌐"
}

func interfaceIPv4Labels(addrsByIface map[string][]net.Addr, name string) []string {
	labels := make([]string, 0)
	for _, addr := range addrsByIface[name] {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			ones, _ := ipnet.Mask.Size()
			labels = append(labels, fmt.Sprintf("%s/%d", ipnet.IP.String(), ones))
		}
	}
	return labels
}
