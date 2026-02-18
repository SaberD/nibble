//go:build darwin

package macos

import (
	"net"
	"net/netip"
	"syscall"

	"github.com/backendsystems/nibble/internal/scan/shared"
	"golang.org/x/net/route"
)

type Neighbor struct {
	IP    string
	MAC   string
	Iface string
}

func LookupMAC(ip string) string {
	for _, row := range Neighbors("") {
		if row.IP == ip && row.MAC != "00:00:00:00:00:00" {
			return row.MAC
		}
	}
	return ""
}

func Neighbors(ifaceName string) []Neighbor {
	rib, err := route.FetchRIB(syscall.AF_INET, route.RIBTypeRoute, 0)
	if err != nil || len(rib) == 0 {
		return nil
	}

	msgs, err := route.ParseRIB(route.RIBTypeRoute, rib)
	if err != nil {
		return nil
	}

	rows := make([]Neighbor, 0)
	seen := make(map[string]struct{})
	for _, msg := range msgs {
		routeMsg, ok := msg.(*route.RouteMessage)
		if !ok || routeMsg.Flags&syscall.RTF_LLINFO == 0 {
			continue
		}

		row, ok := parseNeighbor(routeMsg)
		if !ok {
			continue
		}
		if ifaceName != "" && row.Iface != ifaceName {
			continue
		}

		key := row.IP + "|" + row.Iface
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		rows = append(rows, row)
	}

	return rows
}

func parseNeighbor(routeMsg *route.RouteMessage) (Neighbor, bool) {
	if len(routeMsg.Addrs) <= syscall.RTAX_GATEWAY {
		return Neighbor{}, false
	}

	dst, ok := routeMsg.Addrs[syscall.RTAX_DST].(*route.Inet4Addr)
	if !ok {
		return Neighbor{}, false
	}

	gateway, ok := routeMsg.Addrs[syscall.RTAX_GATEWAY].(*route.LinkAddr)
	if !ok || len(gateway.Addr) == 0 {
		return Neighbor{}, false
	}

	addr := netip.AddrFrom4(dst.IP)
	mac := shared.NormalizeMAC(net.HardwareAddr(gateway.Addr).String())
	if mac == "" {
		return Neighbor{}, false
	}

	iface := gateway.Name
	if iface == "" && routeMsg.Index > 0 {
		netIface, err := net.InterfaceByIndex(routeMsg.Index)
		if err == nil {
			iface = netIface.Name
		}
	}

	return Neighbor{
		IP:    addr.String(),
		MAC:   mac,
		Iface: iface,
	}, true
}
