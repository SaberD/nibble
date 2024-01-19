package scan

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/gosnmp/gosnmp"
)

func PerformScan(subnet string) []string {
	ip, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var wg sync.WaitGroup
	hosts := make([]string, 0)
	var mutex sync.Mutex

	incrementIP := func(ip net.IP) {
		for j := len(ip) - 1; j >= 0; j-- {
			ip[j]++
			if ip[j] > 0 {
				break
			}
		}
	}

	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incrementIP(ip) {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			conn, err := net.DialTimeout("tcp", ip+":80", 2*time.Second)
			if err == nil {
				conn.Close()
				hostInfo := ip

				names, err := net.LookupAddr(ip)
				if err == nil && len(names) > 0 {
					hostInfo += ", " + names[0]
				}

				mutex.Lock()
				hosts = append(hosts, hostInfo)
				mutex.Unlock()
			}
		}(ip.String())
	}

	wg.Wait()
	return hosts
}

// "192.168.1.1", "public", "1.3.6.1.2.1.1.1.0"
func QueryRouterSNMP(target string, community string, oid string) {
	gosnmp.Default.Target = target
	gosnmp.Default.Community = community
	gosnmp.Default.Version = gosnmp.Version2c
	err := gosnmp.Default.Connect()
	if err != nil {
		log.Fatalf("Connect() err: %v", err)
	}
	defer gosnmp.Default.Conn.Close()

	result, err2 := gosnmp.Default.Get([]string{oid}) // Get() accepts up to g.MAX_OIDS
	if err2 != nil {
		log.Fatalf("Get() err: %v", err2)
	}

	for _, variable := range result.Variables {
		fmt.Printf("oid: %s ", variable.Name)

		// the Value of each variable returned by Get() implements
		// interface{}. You could do a type switch...
		switch variable.Type {
		case gosnmp.OctetString:
			fmt.Printf("string: %s\n", string(variable.Value.([]byte)))
		default:
			// ... or often you're just interested in numeric values.
			// ToBigInt() will return the Value as a BigInt, for plugging
			// into your calculations.
			fmt.Printf("number: %d\n", gosnmp.ToBigInt(variable.Value))
		}
	}
}

// srcIP := net.ParseIP("192.168.1.2")
// dstIP := net.ParseIP("192.168.1.3")
// arpPing("eth0", srcIP, dstIP)
func ArpPing(iface string, srcIP net.IP, dstIP net.IP) {
	// Open the device for capturing
	handle, err := pcap.OpenLive(iface, 65536, true, pcap.BlockForever)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// Set filter
	var filter = fmt.Sprintf("arp and host %s", dstIP)
	err = handle.SetBPFFilter(filter)
	if err != nil {
		log.Fatal(err)
	}

	// Create an ARP request packet
	eth := layers.Ethernet{
		SrcMAC:       net.HardwareAddr{0x00, 0x0c, 0x29, 0x48, 0x55, 0xe6},
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}
	arp := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         1, // 1 for request, 2 for reply
		SourceHwAddress:   []byte{0x00, 0x0c, 0x29, 0x48, 0x55, 0xe6},
		SourceProtAddress: []byte(srcIP),
		DstHwAddress:      []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		DstProtAddress:    []byte(dstIP),
	}

	// Send the packet
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{}
	gopacket.SerializeLayers(buf, opts, &eth, &arp)
	if err := handle.WritePacketData(buf.Bytes()); err != nil {
		log.Fatal(err)
	}

	fmt.Println("ARP request sent")
}

// "192.168.1.2", "192.168.1.255"
func sendICMPBroadcast(localAddr string, broadcastAddr string) {
	c, err := icmp.ListenPacket("ip4:icmp", localAddr)
	if err != nil {
		log.Fatalf("ListenPacket err: %v", err)
	}
	defer c.Close()

	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1,
			Data: []byte("HELLO-R-U-THERE"),
		},
	}
	wb, err := wm.Marshal(nil)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := c.WriteTo(wb, &net.IPAddr{IP: net.ParseIP(broadcastAddr)}); err != nil {
		log.Fatalf("WriteTo err: %v", err)
	}
	fmt.Println("ICMP Echo Request sent")
}
