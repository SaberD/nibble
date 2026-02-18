// Package demo provides fake network data for demo recordings.
package demo

import "net"

// GetInterfaces returns fake interfaces used by demo mode.
func GetInterfaces() ([]net.Interface, map[string][]net.Addr, error) {
	specs := []struct {
		name string
		cidr string
	}{
		{name: "eth0", cidr: "192.168.1.100/24"},
		{name: "wlan0", cidr: "10.0.0.50/24"},
		{name: "docker0", cidr: "172.17.0.1/16"},
		{name: "wg0", cidr: "10.8.0.2/24"},
	}

	ifaces := make([]net.Interface, 0, len(specs))
	addrsByIface := make(map[string][]net.Addr, len(specs))
	for _, s := range specs {
		iface, addrs, err := newInterface(s.name, s.cidr)
		if err != nil {
			return nil, nil, err
		}
		ifaces = append(ifaces, iface)
		addrsByIface[iface.Name] = addrs
	}

	return ifaces, addrsByIface, nil
}

func newInterface(name, cidr string) (net.Interface, []net.Addr, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return net.Interface{}, nil, err
	}
	return net.Interface{Name: name}, []net.Addr{ipnet}, nil
}
