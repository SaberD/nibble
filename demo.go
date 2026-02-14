package main

import (
	"net"
)

// GetDemoInterfaces returns fake network interfaces for demo/anonymized recordings
func GetDemoInterfaces() []ifaceInfo {
	_, ipnet1, _ := net.ParseCIDR("192.168.1.100/24")
	_, ipnet2, _ := net.ParseCIDR("10.0.0.50/24")

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

	return []ifaceInfo{
		{iface: iface1, addrs: []net.Addr{ipnet1}},
		{iface: iface2, addrs: []net.Addr{ipnet2}},
	}
}
