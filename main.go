package main

import (
	"fmt"
	"net"
	"os"

	"nibble/internal/scan"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Check for demo mode (for anonymized recordings)
	demoMode := os.Getenv("NIBBLE_DEMO") == "1"

	var ifaces []ifaceInfo

	if demoMode {
		// Use fake network interfaces for demo
		ifaces = GetDemoInterfaces()
	} else {
		// Get real network interfaces
		interfaces, err := net.Interfaces()
		if err != nil {
			fmt.Println("Error getting network interfaces:", err)
			os.Exit(1)
		}

		ifaces = getRealInterfaces(interfaces)
	}

	if len(ifaces) == 0 {
		fmt.Println("No valid network interfaces found with IPv4 addresses")
		os.Exit(1)
	}

	// Create the appropriate scanner for real or demo mode
	var scanner scan.Scanner
	if demoMode {
		scanner = &scan.DemoScanner{}
	} else {
		scanner = &scan.NetScanner{}
	}

	// Initialize the model
	initialModel := model{
		interfaces: ifaces,
		scanner:    scanner,
		cursor:     0,
		progress: progress.New(
			progress.WithScaledGradient("#FFD700", "#B8B000"), // Bright yellow to muted gold
		),
		selected: false,
	}

	// Start the Bubble Tea program (use alt screen for selection UI)
	prog := tea.NewProgram(initialModel, tea.WithAltScreen())
	finalModel, err := prog.Run()
	if err != nil {
		fmt.Printf("Error starting the program: %v", err)
		os.Exit(1)
	}

	// Print final view so results stay in scrollback after alt screen exits
	if m, ok := finalModel.(model); ok && m.scanComplete {
		fmt.Print(m.View())
	}
}

// getRealInterfaces extracts valid IPv4 network interfaces
func getRealInterfaces(interfaces []net.Interface) []ifaceInfo {
	var ifaces []ifaceInfo
	for _, iface := range interfaces {
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

		ifaces = append(ifaces, ifaceInfo{iface: iface, addrs: addrs})
	}
	return ifaces
}
