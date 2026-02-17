// Package demo contains fake interface data used only for creating demo recording gif using bubble tea vhs
package demo

import "net"

type InterfaceInfo struct {
	Iface net.Interface
	Addrs []net.Addr
}

// GetInterfaces returns fake network interfaces for demo/anonymized recordings.
func GetInterfaces() []InterfaceInfo {
	ipnet1 := &net.IPNet{IP: net.ParseIP("192.168.1.100").To4(), Mask: net.CIDRMask(24, 32)}
	ipnet2 := &net.IPNet{IP: net.ParseIP("10.0.0.50").To4(), Mask: net.CIDRMask(24, 32)}
	ipnet3 := &net.IPNet{IP: net.ParseIP("172.17.0.1").To4(), Mask: net.CIDRMask(16, 32)}
	ipnet4 := &net.IPNet{IP: net.ParseIP("10.8.0.2").To4(), Mask: net.CIDRMask(24, 32)}

	iface1 := net.Interface{
		Index:        2,
		MTU:          1500,
		Name:         "eth0",
		HardwareAddr: net.HardwareAddr{0x08, 0x00, 0x27, 0x00, 0x00, 0x00},
		Flags:        net.FlagUp | net.FlagBroadcast | net.FlagRunning | net.FlagMulticast,
	}

	iface2 := net.Interface{
		Index:        3,
		MTU:          1500,
		Name:         "wlan0",
		HardwareAddr: net.HardwareAddr{0x08, 0x00, 0x27, 0x00, 0x00, 0x01},
		Flags:        net.FlagUp | net.FlagBroadcast | net.FlagRunning | net.FlagMulticast,
	}

	iface3 := net.Interface{
		Index:        4,
		MTU:          1500,
		Name:         "docker0",
		HardwareAddr: net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x01},
		Flags:        net.FlagUp | net.FlagBroadcast | net.FlagRunning | net.FlagMulticast,
	}

	iface4 := net.Interface{
		Index:        5,
		MTU:          1420,
		Name:         "wg0",
		HardwareAddr: net.HardwareAddr{},
		Flags:        net.FlagUp | net.FlagPointToPoint | net.FlagRunning,
	}

	return []InterfaceInfo{
		{Iface: iface1, Addrs: []net.Addr{ipnet1}},
		{Iface: iface2, Addrs: []net.Addr{ipnet2}},
		{Iface: iface3, Addrs: []net.Addr{ipnet3}},
		{Iface: iface4, Addrs: []net.Addr{ipnet4}},
	}
}
