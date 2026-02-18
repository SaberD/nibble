package tui

import (
	"net"

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
	return nil
}
