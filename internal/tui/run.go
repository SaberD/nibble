package tui

import (
	"net"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scan"
	"github.com/backendsystems/nibble/internal/scanner"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

func Run(networkScanner scanner.Scanner, ifaces []net.Interface, addrsByIface map[string][]net.Addr) error {
	cfg, _ := ports.LoadConfig()
	pack := cfg.Mode
	if pack == "" || !ports.IsValidPack(pack) {
		pack = "default"
	}
	if netScanner, ok := networkScanner.(*scan.NetScanner); ok {
		if resolvedPorts, err := ports.Resolve(pack, cfg.Custom, ""); err == nil {
			netScanner.Ports = resolvedPorts
		}
	}

	initialModel := model{
		interfaces:   ifaces,
		addrsByIface: addrsByIface,
		scanner:      networkScanner,
		cursor:       0,
		editingPorts: false,
		portPack:     pack,
		customPorts:  cfg.Custom,
		customCursor: len(cfg.Custom),
		progress: progress.New(
			progress.WithScaledGradient("#FFD700", "#B8B000"),
		),
	}

	prog := tea.NewProgram(initialModel)
	_, err := prog.Run()
	return err
}
