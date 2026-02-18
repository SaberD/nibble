package tui

import "strings"

func interfaceIcon(name string) string {
	lower := strings.ToLower(name)

	// Container/virtual network interfaces (Docker, Podman, Kubernetes/CNI, LXC/LXD, libvirt).
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

	// Common VPN/tunnel interface prefixes across Linux/macOS/Windows.
	if strings.HasPrefix(lower, "tun") ||
		strings.HasPrefix(lower, "tap") ||
		strings.HasPrefix(lower, "utun") ||
		strings.HasPrefix(lower, "wg") ||
		strings.HasPrefix(lower, "tailscale") ||
		strings.Contains(lower, "vpn") {
		return "🔒"
	}

	// Wi-Fi adapters: Linux-style (wl*/wlan*) and Windows/macOS naming.
	if strings.HasPrefix(lower, "wl") ||
		strings.HasPrefix(lower, "wlan") ||
		strings.Contains(lower, "wi-fi") ||
		strings.Contains(lower, "wifi") ||
		strings.Contains(lower, "wireless") {
		return "📶"
	}

	// Ethernet adapters: Linux-style (en*/eth*) and Windows naming.
	if strings.HasPrefix(lower, "en") ||
		strings.HasPrefix(lower, "eth") ||
		strings.Contains(lower, "ethernet") {
		return "🔌"
	}
	return "🌐"
}
