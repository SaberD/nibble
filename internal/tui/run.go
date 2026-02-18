package tui

import (
	"net"

	"github.com/backendsystems/nibble/internal/scanner"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

func Run(networkScanner scanner.Scanner, ifaces []net.Interface, addrsByIface map[string][]net.Addr) error {
	initialModel := model{
		interfaces:   ifaces,
		addrsByIface: addrsByIface,
		scanner:      networkScanner,
		cursor:       0,
		progress: progress.New(
			progress.WithScaledGradient("#FFD700", "#B8B000"),
		),
	}

	prog := tea.NewProgram(initialModel)
	_, err := prog.Run()
	return err
}
