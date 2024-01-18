package main

import (
	"fmt"
	"net"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func main() {
	// Get network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Error getting network interfaces:", err)
		os.Exit(1)
	}

	var ifaces []ifaceInfo
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Println("Error getting addresses:", err)
			continue // or handle error appropriately
		}
		ifaces = append(ifaces, ifaceInfo{iface: iface, addrs: addrs})
	}

	// Initialize the model
	initialModel := model{
		ifaces:   ifaces,
		cursor:   0,
		selected: false,
	}

	// Start the Bubble Tea program
	p := tea.NewProgram(initialModel)
	if err := p.Start(); err != nil {
		fmt.Printf("Error starting the program: %v", err)
		os.Exit(1)
	}

	if initialModel.selected {
		fmt.Printf("You selected: %s\n", initialModel.selectedIface.iface.Name)
		for _, addr := range initialModel.selectedIface.addrs {
			fmt.Printf("IP Address: %s\n", addr.String())
		}
	}
}

func ping(addr string) {
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		fmt.Println("Error listening for ICMP packets:", err)
		return
	}
	defer c.Close()

	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1, //<<16 | 1,
			Data: []byte("HELLO-R-U-THERE"),
		},
	}
	wb, err := wm.Marshal(nil)
	if err != nil {
		fmt.Println("Error marshalling message:", err)
		return
	}

	if _, err := c.WriteTo(wb, &net.IPAddr{IP: net.ParseIP(addr)}); err != nil {
		fmt.Println("Error writing to buffer:", err)
		return
	}

	c.SetReadDeadline(time.Now().Add(3 * time.Second))
	rb := make([]byte, 1500)
	n, _, err := c.ReadFrom(rb)
	if err != nil {
		return // Assume that no response means the host is down or unreachable.
	}
	rm, err := icmp.ParseMessage(1, rb[:n])
	if err != nil {
		fmt.Println("Error parsing ICMP message:", err)
		return
	}

	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		fmt.Printf("Host is up: %s\n", addr)
	default:
		// Not an echo reply, ignore it
	}
}
