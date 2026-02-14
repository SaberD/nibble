package scan

import (
	"encoding/binary"
	"net"
	"strings"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

type mdnsService struct {
	IP       string
	Instance string
	Hostname string
	Port     uint16
}

func discoverSSHViaMDNS(ifaceName string, subnet *net.IPNet) []mdnsService {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return nil
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil
	}

	var ifaceIPv4 net.IP
	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if ip := ipnet.IP.To4(); ip != nil {
			ifaceIPv4 = ip
			break
		}
	}
	if ifaceIPv4 == nil {
		return nil
	}

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: ifaceIPv4, Port: 0})
	if err != nil {
		return nil
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(1200 * time.Millisecond)); err != nil {
		return nil
	}

	queryName, err := dnsmessage.NewName("_ssh._tcp.local.")
	if err != nil {
		return nil
	}

	var idBytes [2]byte
	binary.BigEndian.PutUint16(idBytes[:], uint16(time.Now().UnixNano()))
	query := dnsmessage.Message{
		Header: dnsmessage.Header{
			ID:                 binary.BigEndian.Uint16(idBytes[:]),
			RecursionDesired:   false,
			Response:           false,
			Authoritative:      false,
			RecursionAvailable: false,
		},
		Questions: []dnsmessage.Question{
			{
				Name:  queryName,
				Type:  dnsmessage.TypePTR,
				Class: dnsmessage.ClassINET,
			},
		},
	}

	packet, err := query.Pack()
	if err != nil {
		return nil
	}

	mdnsAddr := &net.UDPAddr{IP: net.IPv4(224, 0, 0, 251), Port: 5353}
	for i := 0; i < 2; i++ {
		_, _ = conn.WriteToUDP(packet, mdnsAddr)
	}

	type srvInfo struct {
		target string
		port   uint16
	}

	ptrInstances := make(map[string]struct{})
	srvByInstance := make(map[string]srvInfo)
	ipByHost := make(map[string][]net.IP)

	recordIP := func(host string, ip net.IP) {
		if ip == nil {
			return
		}
		ipByHost[host] = append(ipByHost[host], ip)
	}

	parseResources := func(resources []dnsmessage.Resource) {
		for _, r := range resources {
			name := normalizeMDNSName(r.Header.Name.String())

			switch body := r.Body.(type) {
			case *dnsmessage.PTRResource:
				if name != "_ssh._tcp.local" {
					continue
				}
				ptrInstances[normalizeMDNSName(body.PTR.String())] = struct{}{}
			case *dnsmessage.SRVResource:
				srvByInstance[name] = srvInfo{
					target: normalizeMDNSName(body.Target.String()),
					port:   body.Port,
				}
			case *dnsmessage.AResource:
				recordIP(name, net.IPv4(body.A[0], body.A[1], body.A[2], body.A[3]))
			case *dnsmessage.AAAAResource:
				recordIP(name, net.IP(body.AAAA[:]))
			}
		}
	}

	buf := make([]byte, 2048)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				break
			}
			break
		}

		var msg dnsmessage.Message
		if err := msg.Unpack(buf[:n]); err != nil {
			continue
		}

		parseResources(msg.Answers)
		parseResources(msg.Additionals)
	}

	seen := make(map[string]struct{})
	var services []mdnsService
	for instance := range ptrInstances {
		srv, ok := srvByInstance[instance]
		if !ok {
			continue
		}

		for _, ip := range ipByHost[srv.target] {
			ipv4 := ip.To4()
			if ipv4 == nil || !subnet.Contains(ipv4) {
				continue
			}

			ipStr := ipv4.String()
			if _, exists := seen[ipStr]; exists {
				continue
			}
			seen[ipStr] = struct{}{}

			services = append(services, mdnsService{
				IP:       ipStr,
				Instance: instance,
				Hostname: srv.target,
				Port:     srv.port,
			})
		}
	}

	return services
}

func normalizeMDNSName(name string) string {
	name = strings.TrimSpace(name)
	return strings.TrimSuffix(name, ".")
}
