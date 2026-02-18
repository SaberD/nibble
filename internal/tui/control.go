package tui

import (
	"fmt"
	"net"

	"github.com/backendsystems/nibble/internal/scanner"

	tea "github.com/charmbracelet/bubbletea"
)

type scanProgressMsg struct {
	update scanner.ProgressUpdate
}
type scanCompleteMsg struct{}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width

	case tea.KeyMsg:
		return m.handleKey(msg)

	case scanProgressMsg:
		return m.handleScanProgress(msg)

	case scanCompleteMsg:
		m.scanning = false
		m.scanComplete = true
		return m, tea.Quit
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.scanning || m.scanComplete {
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			// Preserve current results when quitting mid-scan so the final
			// rendered view remains in scrollback after alt-screen exits.
			if m.scanning {
				m.scanning = false
				m.scanComplete = true
			}
			return m, tea.Quit
		}
		return m, nil
	}

	// Close help on any key when it's shown.
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "?":
		m.showHelp = true
		return m, nil

	case "up", "w", "k":
		m.moveCursorUp()

	case "down", "s", "j":
		m.moveCursorDown()

	case "left", "a", "h":
		if m.cursor > 0 {
			m.cursor--
		}

	case "right", "d", "l":
		if m.cursor < len(m.interfaces)-1 {
			m.cursor++
		}

	case "enter":
		return m.selectInterfaceAndScan()
	}

	return m, nil
}

func (m model) moveCursorUp() {
	cardsPerRow := m.cardsPerRow()
	if m.cursor >= cardsPerRow {
		m.cursor -= cardsPerRow
	}
}

func (m model) moveCursorDown() {
	cardsPerRow := m.cardsPerRow()
	if m.cursor+cardsPerRow < len(m.interfaces) {
		m.cursor += cardsPerRow
	}
}

func (m model) cardsPerRow() int {
	cardWidth := 26 // 22 + 2 for border + 2 for spacing
	cardsPerRow := (m.windowWidth - 4) / cardWidth
	if cardsPerRow < 1 {
		return 1
	}
	return cardsPerRow
}

func (m model) selectInterfaceAndScan() (tea.Model, tea.Cmd) {
	if m.cursor >= len(m.interfaces) {
		return m, nil
	}

	m.selectedIface = m.interfaces[m.cursor]
	m.selectedAddrs = m.addrsByIface[m.selectedIface.Name]

	if len(m.selectedAddrs) == 0 {
		m.errorMsg = fmt.Sprintf("interface %s has no IP addresses", m.selectedIface.Name)
		return m, nil
	}

	targetAddr := scanner.FirstIp4(m.selectedAddrs)
	if targetAddr == "" {
		m.errorMsg = fmt.Sprintf("interface %s has no valid IPv4 addresses", m.selectedIface.Name)
		return m, nil
	}

	// Calculate total hosts for progress.
	_, ipnet, _ := net.ParseCIDR(targetAddr)
	m.totalHosts = scanner.TotalScanHosts(ipnet)

	m.scanning = true
	m.errorMsg = ""
	m.foundHosts = nil
	m.scannedCount = 0
	m.neighborSeen = 0
	m.neighborTotal = 0
	m.progressChan = make(chan scanner.ProgressUpdate, 256)

	return m, performScan(m.scanner, m.selectedIface.Name, targetAddr, m.progressChan)
}

func (m model) handleScanProgress(msg scanProgressMsg) (tea.Model, tea.Cmd) {
	switch p := msg.update.(type) {
	case scanner.NeighborProgress:
		if p.TotalHosts > 0 {
			m.totalHosts = p.TotalHosts
		}
		m.neighborSeen = p.Seen
		m.neighborTotal = p.Total
		if p.Host != "" {
			m.foundHosts = append(m.foundHosts, p.Host)
		}
	case scanner.SweepProgress:
		if p.TotalHosts > 0 {
			m.totalHosts = p.TotalHosts
		}
		m.scannedCount = p.Scanned
		if p.Host != "" {
			m.foundHosts = append(m.foundHosts, p.Host)
		}
	}
	return m, listenForProgress(m.progressChan)
}

func listenForProgress(progressChan <-chan scanner.ProgressUpdate) tea.Cmd {
	return func() tea.Msg {
		progress, ok := <-progressChan
		if !ok {
			return scanCompleteMsg{}
		}
		return scanProgressMsg{update: progress}
	}
}

func performScan(networkScanner scanner.Scanner, ifaceName string, targetAddr string, progressChan chan scanner.ProgressUpdate) tea.Cmd {
	return func() tea.Msg {
		go networkScanner.ScanNetwork(ifaceName, targetAddr, progressChan)
		return listenForProgress(progressChan)()
	}
}
