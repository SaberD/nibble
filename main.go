package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/backendsystems/nibble/internal/demo"
	"github.com/backendsystems/nibble/internal/scan"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Enable demo mode for anonymized recordings.
	var demoMode bool
	flag.BoolVar(&demoMode, "demo", false, "use demo interfaces")
	flag.Parse()

	var ifaces []net.Interface
	var addrsByIface map[string][]net.Addr

	if demoMode {
		// Use fake network interfaces for demo
		var err error
		ifaces, addrsByIface, err = demo.GetInterfaces()
		if err != nil {
			fmt.Println("Error creating demo interfaces:", err)
			os.Exit(1)
		}
	} else {
		// Get real network interfaces
		sysIfaces, err := net.Interfaces()
		if err != nil {
			fmt.Println("Error getting network interfaces:", err)
			os.Exit(1)
		}

		ifaces, addrsByIface = getRealInterfaces(sysIfaces)
	}

	if len(ifaces) == 0 {
		fmt.Println("No valid network interfaces found with IPv4 addresses")
		os.Exit(1)
	}

	var scanner scan.Scanner
	if demoMode {
		scanner = &scan.DemoScanner{}
	} else {
		scanner = &scan.NetScanner{}
	}

	// Initialize the model
	initialModel := model{
		interfaces:   ifaces,
		addrsByIface: addrsByIface,
		scanner:      scanner,
		cursor:       0,
		progress: progress.New(
			progress.WithScaledGradient("#FFD700", "#B8B000"), // Bright yellow to muted gold
		),
		selected: false,
	}

	// Start Bubble Tea in the normal terminal screen so the final scan
	// output remains visible and the shell prompt returns directly below it.
	prog := tea.NewProgram(initialModel)
	_, err := prog.Run()
	if err != nil {
		fmt.Printf("Error starting the program: %v", err)
		os.Exit(1)
	}
}

// getRealInterfaces extracts valid IPv4 network interfaces
func getRealInterfaces(sysIfaces []net.Interface) ([]net.Interface, map[string][]net.Addr) {
	ifaces := make([]net.Interface, 0, len(sysIfaces))
	addrsByIface := make(map[string][]net.Addr, len(sysIfaces))
	for _, iface := range sysIfaces {
		// Skip loopback interfaces
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Skip interfaces that are down
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		// Skip interfaces with no addresses
		if len(addrs) == 0 {
			continue
		}

		// Check for at least one IPv4 address
		hasIPv4 := false
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipnet.IP.To4() != nil {
					hasIPv4 = true
					break
				}
			}
		}

		if !hasIPv4 {
			continue
		}

		ifaces = append(ifaces, iface)
		addrsByIface[iface.Name] = addrs
	}
	return ifaces, addrsByIface
}
