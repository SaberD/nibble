package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/backendsystems/nibble/internal/demo"
	"github.com/backendsystems/nibble/internal/scan"
	"github.com/backendsystems/nibble/internal/scanner"
	"github.com/backendsystems/nibble/internal/tui"
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
		// Get real network interfaces.
		var err error
		ifaces, addrsByIface, err = scan.DiscoverInterfaces()
		if err != nil {
			fmt.Println("Error getting network interfaces:", err)
			os.Exit(1)
		}
	}

	if len(ifaces) == 0 {
		fmt.Println("No valid network interfaces found with IPv4 addresses")
		os.Exit(1)
	}

	var networkScanner scanner.Scanner
	if demoMode {
		networkScanner = &demo.DemoScanner{}
	} else {
		networkScanner = &scan.NetScanner{}
	}

	if err := tui.Run(networkScanner, ifaces, addrsByIface); err != nil {
		fmt.Printf("Error starting the program: %v", err)
		os.Exit(1)
	}
}
