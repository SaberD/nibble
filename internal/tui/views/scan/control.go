package scanview

import (
	"net"
	"strings"

	"github.com/backendsystems/nibble/internal/scanner"

	tea "github.com/charmbracelet/bubbletea"
)

const scanHelpText = "j/k or ↑/↓: scroll • q: quit"

// appendIfNew appends host to hosts only if no existing entry has the same IP.
func appendIfNew(hosts []string, host string) []string {
	newIP := hostIP(host)
	for _, h := range hosts {
		if hostIP(h) == newIP {
			return hosts
		}
	}
	return append(hosts, host)
}

// hostIP extracts the IP address from the first line of a host string.
// Host format is "IP - vendor" or just "IP".
func hostIP(host string) string {
	line, _, _ := strings.Cut(host, "\n")
	ip, _, _ := strings.Cut(line, " - ")
	return strings.TrimSpace(ip)
}

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
type QuitMsg struct{}

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
	m.ShouldPrintFinal = false
	m.FoundHosts = nil
	m.FinalHosts = nil
	m.ScannedCount = 0
	m.NeighborSeen = 0
	m.NeighborTotal = 0
	m.ProgressChan = make(chan scanner.ProgressUpdate, 256)
	m = m.RefreshResults(false)
	return m, PerformScan(m.NetworkScan, iface.Name, targetAddr, m.ProgressChan)
}

func (m Model) Update(msg tea.Msg) Result {
	result := Result{Model: m}
	switch typed := msg.(type) {
	case tea.KeyMsg:
		result.Handled = true
		switch HandleKey(m.Scanning, m.ScanComplete, typed.String()) {
		case ActionQuitAndComplete:
			result.Model = prepareForExit(result.Model, true)
			result.Model.Scanning = false
			result.Model.ScanComplete = true
			result.Cmd = sendQuitMsg()
			return result
		case ActionQuit:
			result.Cmd = sendQuitMsg()
			return result
		}
		var cmd tea.Cmd
		result.Model.Results, cmd = m.Results.Update(typed)
		result.Cmd = cmd
		return result
	case ProgressMsg:
		result.Handled = true
		hostAdded := false
		switch p := typed.Update.(type) {
		case scanner.NeighborProgress:
			if p.TotalHosts > 0 {
				result.Model.TotalHosts = p.TotalHosts
			}
			result.Model.NeighborSeen = p.Seen
			result.Model.NeighborTotal = p.Total
			if p.Host != "" {
				before := len(result.Model.FoundHosts)
				result.Model.FoundHosts = appendIfNew(result.Model.FoundHosts, p.Host)
				hostAdded = len(result.Model.FoundHosts) > before
			}
		case scanner.SweepProgress:
			if p.TotalHosts > 0 {
				result.Model.TotalHosts = p.TotalHosts
			}
			result.Model.ScannedCount = p.Scanned
			if p.Host != "" {
				before := len(result.Model.FoundHosts)
				result.Model.FoundHosts = appendIfNew(result.Model.FoundHosts, p.Host)
				hostAdded = len(result.Model.FoundHosts) > before
			}
		}
		if hostAdded {
			result.Model = result.Model.RefreshResults(true)
		}
		result.Cmd = ListenForProgress(m.ProgressChan)
		return result
	case CompleteMsg:
		result.Handled = true
		result.Model = prepareForExit(result.Model, true)
		result.Model.Scanning = false
		result.Model.ScanComplete = true
		result.Cmd = sendQuitMsg()
		return result
	case QuitMsg:
		result.Handled = true
		result.Quit = true
		return result
	default:
		return result
	}
}

func sendQuitMsg() tea.Cmd {
	return func() tea.Msg { return QuitMsg{} }
}

// prepareForExit preserves discovered hosts for optional final output, then clears
// the live viewport data to avoid duplicate terminal content on exit.
func prepareForExit(m Model, shouldPrint bool) Model {
	m.ShouldPrintFinal = shouldPrint
	if len(m.FinalHosts) == 0 && len(m.FoundHosts) > 0 {
		m.FinalHosts = append([]string(nil), m.FoundHosts...)
	}
	m.FoundHosts = nil
	m.Results.SetContent("")
	return m
}
