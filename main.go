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
			progress.WithScaledGradient("#FFD700", "#B8B000"),
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
