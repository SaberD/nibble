package scanview

import (
	"net"

	"github.com/backendsystems/nibble/internal/scanner"

	tea "github.com/charmbracelet/bubbletea"
)

const scanHelpText = "q: quit"

type Action int

const (
	ActionNone Action = iota
	ActionQuit
	ActionQuitAndComplete
)

type ProgressMsg struct {
	Update scanner.ProgressUpdate
}

type CompleteMsg struct{}

type Result struct {
	Model   Model
	Handled bool
	Quit    bool
	Cmd     tea.Cmd
}

func HandleKey(scanning bool, scanComplete bool, key string) Action {
	if !scanning && !scanComplete {
		return ActionNone
	}
	if key != "ctrl+c" && key != "q" {
		return ActionNone
	}
	if scanning {
		return ActionQuitAndComplete
	}
	return ActionQuit
}

func ListenForProgress(progressChan <-chan scanner.ProgressUpdate) tea.Cmd {
	return func() tea.Msg {
		progress, ok := <-progressChan
		if !ok {
			return CompleteMsg{}
		}
		return ProgressMsg{Update: progress}
	}
}

func PerformScan(networkScanner scanner.Scanner, ifaceName, targetAddr string, progressChan chan scanner.ProgressUpdate) tea.Cmd {
	return func() tea.Msg {
		go networkScanner.ScanNetwork(ifaceName, targetAddr, progressChan)
		return ListenForProgress(progressChan)()
	}
}

func (m Model) Start(iface net.Interface, addrs []net.Addr, totalHosts int, targetAddr string) (Model, tea.Cmd) {
	m.SelectedIface = iface
	m.SelectedAddrs = addrs
	m.TotalHosts = totalHosts
	m.Scanning = true
	m.ScanComplete = false
	m.FoundHosts = nil
	m.ScannedCount = 0
	m.NeighborSeen = 0
	m.NeighborTotal = 0
	m.ProgressChan = make(chan scanner.ProgressUpdate, 256)
	return m, PerformScan(m.NetworkScan, iface.Name, targetAddr, m.ProgressChan)
}

func (m Model) Update(msg tea.Msg) Result {
	result := Result{Model: m}
	switch typed := msg.(type) {
	case tea.KeyMsg:
		result.Handled = true
		switch HandleKey(m.Scanning, m.ScanComplete, typed.String()) {
		case ActionQuitAndComplete:
			result.Model.Scanning = false
			result.Model.ScanComplete = true
			result.Quit = true
		case ActionQuit:
			result.Quit = true
		}
		return result
	case ProgressMsg:
		result.Handled = true
		switch p := typed.Update.(type) {
		case scanner.NeighborProgress:
			if p.TotalHosts > 0 {
				result.Model.TotalHosts = p.TotalHosts
			}
			result.Model.NeighborSeen = p.Seen
			result.Model.NeighborTotal = p.Total
			if p.Host != "" {
				result.Model.FoundHosts = append(result.Model.FoundHosts, p.Host)
			}
		case scanner.SweepProgress:
			if p.TotalHosts > 0 {
				result.Model.TotalHosts = p.TotalHosts
			}
			result.Model.ScannedCount = p.Scanned
			if p.Host != "" {
				result.Model.FoundHosts = append(result.Model.FoundHosts, p.Host)
			}
		}
		result.Cmd = ListenForProgress(m.ProgressChan)
		return result
	case CompleteMsg:
		result.Handled = true
		result.Model.Scanning = false
		result.Model.ScanComplete = true
		result.Quit = true
		return result
	default:
		return result
	}
}
