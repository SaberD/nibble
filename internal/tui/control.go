package tui

import (
	"fmt"
	"net"
	"strings"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scan"
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

	if m.editingPorts {
		return m.handlePortsKey(msg)
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

	case "p":
		m.editingPorts = true
		m.customCursor = len(m.customPorts)
		return m, nil

	case "w", "k":
		m.moveCursorUp()

	case "s", "j":
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

func (m model) handlePortsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "tab", "up", "down":
		if m.portPack == "default" {
			m.portPack = "custom"
		} else {
			m.portPack = "default"
		}
		if m.customCursor > len(m.customPorts) {
			m.customCursor = len(m.customPorts)
		}
		return m, nil
	case "enter":
		return m.applyPortConfigAndContinue()
	case "left", "h":
		if m.portPack == "custom" && m.customCursor > 0 {
			m.customCursor--
		}
		return m, nil
	case "right", "l":
		if m.portPack == "custom" && m.customCursor < len(m.customPorts) {
			m.customCursor++
		}
		return m, nil
	case "home", "ctrl+a":
		if m.portPack == "custom" {
			m.customCursor = 0
		}
		return m, nil
	case "end", "ctrl+e":
		if m.portPack == "custom" {
			m.customCursor = len(m.customPorts)
		}
		return m, nil
	case "backspace":
		if m.portPack != "custom" || m.customCursor == 0 || len(m.customPorts) == 0 {
			return m, nil
		}
		i := m.customCursor - 1
		m.customPorts = m.customPorts[:i] + m.customPorts[m.customCursor:]
		m.customCursor--
		return m, nil
	case "delete":
		if m.portPack != "custom" {
			return m, nil
		}
		m.customPorts = ""
		m.customCursor = 0
		return m, nil
	}

	if m.portPack != "custom" {
		return m, nil
	}

	if msg.Type == tea.KeyRunes {
		for _, r := range msg.Runes {
			if r >= '0' && r <= '9' {
				if !canInsertPortDigit(m.customPorts, m.customCursor, r) {
					continue
				}
				s := string(r)
				m.customPorts = m.customPorts[:m.customCursor] + s + m.customPorts[m.customCursor:]
				m.customCursor++
				continue
			}
			if r == ',' {
				s := string(r)
				m.customPorts = m.customPorts[:m.customCursor] + s + m.customPorts[m.customCursor:]
				m.customCursor++
			}
		}
	}
	return m, nil
}

func (m model) applyPortConfigAndContinue() (tea.Model, tea.Cmd) {
	addPorts := ""
	if m.portPack == "custom" {
		normalized, normErr := ports.NormalizeCustom(strings.TrimSpace(m.customPorts))
		if normErr != nil {
			m.errorMsg = normErr.Error()
			return m, nil
		}
		addPorts = normalized
		m.customPorts = normalized
		m.customCursor = len(m.customPorts)
	}

	resolvedPorts, err := ports.Resolve(m.portPack, addPorts, "")
	if err != nil {
		m.errorMsg = err.Error()
		return m, nil
	}

	if cfgErr := ports.SaveConfig(ports.Config{
		Mode:   m.portPack,
		Custom: addPorts,
	}); cfgErr != nil {
		m.errorMsg = cfgErr.Error()
		return m, nil
	}

	if netScanner, ok := m.scanner.(*scan.NetScanner); ok {
		netScanner.Ports = resolvedPorts
	}

	m.errorMsg = ""
	m.editingPorts = false
	return m, nil
}

func canInsertPortDigit(s string, cursor int, digit rune) bool {
	start, end := currentTokenBounds(s, cursor)
	pos := cursor - start
	token := s[start:end]
	next := token[:pos] + string(digit) + token[pos:]

	return len(next) <= 5
}

func currentTokenBounds(s string, cursor int) (int, int) {
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(s) {
		cursor = len(s)
	}

	start := strings.LastIndexByte(s[:cursor], ',')
	if start == -1 {
		start = 0
	} else {
		start++
	}

	rest := s[cursor:]
	nextComma := strings.IndexByte(rest, ',')
	end := len(s)
	if nextComma >= 0 {
		end = cursor + nextComma
	}

	return start, end
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
