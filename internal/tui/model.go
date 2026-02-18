package tui

import (
	"net"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scanner"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	progress      progress.Model
	interfaces    []net.Interface
	addrsByIface  map[string][]net.Addr
	scanner       scanner.Scanner
	cursor        int
	selectedIface net.Interface
	selectedAddrs []net.Addr
	errorMsg      string
	showHelp      bool
	editingPorts  bool
	portPack      string
	customPorts   string
	customCursor  int
	portConfigLoc string
	scanning      bool
	scanComplete  bool
	foundHosts    []string
	scannedCount  int
	totalHosts    int
	neighborSeen  int
	neighborTotal int
	progressChan  chan scanner.ProgressUpdate
	windowWidth   int
}

func (m model) Init() tea.Cmd {
	if m.portPack == "" {
		m.portPack = "default"
	}
	if m.portConfigLoc == "" {
		if path, err := ports.ConfigPath(); err == nil {
			m.portConfigLoc = path
		}
	}
	if m.customCursor < 0 || m.customCursor > len(m.customPorts) {
		m.customCursor = len(m.customPorts)
	}
	return nil
}
