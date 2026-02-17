package tui

import (
	"fmt"
	"net"
	"strings"

	"github.com/backendsystems/nibble/internal/scan"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func Run(scanner scan.Scanner, ifaces []net.Interface, addrsByIface map[string][]net.Addr) error {
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

	prog := tea.NewProgram(initialModel)
	_, err := prog.Run()
	return err
}

func interfaceIcon(name string) string {
	lower := strings.ToLower(name)

	// Container/virtual network interfaces (Docker, Podman, Kubernetes/CNI, LXC/LXD, libvirt).
	if strings.HasPrefix(lower, "docker") ||
		strings.HasPrefix(lower, "br-") ||
		strings.HasPrefix(lower, "veth") ||
		strings.HasPrefix(lower, "cni") ||
		strings.HasPrefix(lower, "flannel") ||
		strings.HasPrefix(lower, "cali") ||
		strings.HasPrefix(lower, "virbr") ||
		strings.HasPrefix(lower, "lxc") ||
		strings.HasPrefix(lower, "podman") {
		return "📦"
	}

	// Common VPN/tunnel interface prefixes across Linux/macOS/Windows.
	if strings.HasPrefix(lower, "tun") ||
		strings.HasPrefix(lower, "tap") ||
		strings.HasPrefix(lower, "utun") ||
		strings.HasPrefix(lower, "wg") ||
		strings.HasPrefix(lower, "tailscale") ||
		strings.Contains(lower, "vpn") {
		return "🔒"
	}

	// Wi-Fi adapters: Linux-style (wl*/wlan*) and Windows/macOS naming.
	if strings.HasPrefix(lower, "wl") ||
		strings.HasPrefix(lower, "wlan") ||
		strings.Contains(lower, "wi-fi") ||
		strings.Contains(lower, "wifi") ||
		strings.Contains(lower, "wireless") {
		return "📶"
	}

	// Ethernet adapters: Linux-style (en*/eth*) and Windows naming.
	if strings.HasPrefix(lower, "en") ||
		strings.HasPrefix(lower, "eth") ||
		strings.Contains(lower, "ethernet") {
		return "🔌"
	}
	return "🌐"
}

type model struct {
	progress      progress.Model
	interfaces    []net.Interface
	addrsByIface  map[string][]net.Addr
	scanner       scan.Scanner
	cursor        int
	selectedIface net.Interface
	selectedAddrs []net.Addr
	selected      bool
	errorMsg      string
	showHelp      bool
	scanning      bool
	scanComplete  bool
	foundHosts    []string
	scannedCount  int
	totalHosts    int
	neighborSeen  int
	neighborTotal int
	progressChan  chan scan.ScanProgress
	windowWidth   int
	windowHeight  int
}

type scanProgressMsg scan.ScanProgress
type scanCompleteMsg struct{}

func (m model) Init() tea.Cmd {
	return nil
}

func listenForProgress(progressChan <-chan scan.ScanProgress) tea.Cmd {
	return func() tea.Msg {
		progress, ok := <-progressChan
		if !ok {
			return scanCompleteMsg{}
		}
		return scanProgressMsg(progress)
	}
}

func totalScanHosts(ipnet *net.IPNet) int {
	ones, bits := ipnet.Mask.Size()
	hostBits := bits - ones

	// Non-IPv4 fallback keeps prior behavior.
	if bits != 32 {
		return 1 << uint(hostBits)
	}

	switch {
	case hostBits <= 0:
		return 1
	case hostBits == 1:
		return 2
	default:
		return (1 << uint(hostBits)) - 2
	}
}

func performScan(scanner scan.Scanner, ifaceName string, targetAddr string, progressChan chan scan.ScanProgress) tea.Cmd {
	return func() tea.Msg {
		go scanner.ScanNetwork(ifaceName, targetAddr, progressChan)
		return listenForProgress(progressChan)()
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

	case tea.KeyMsg:
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

		// Close help on any key when it's shown
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "?":
			// Toggle help overlay
			m.showHelp = true
			return m, nil

		case "up", "w", "k":
			// Calculate cards per row for navigation
			cardWidth := 26 // 22 + 2 for border + 2 for spacing
			cardsPerRow := (m.windowWidth - 4) / cardWidth
			if cardsPerRow < 1 {
				cardsPerRow = 1
			}
			if m.cursor >= cardsPerRow {
				m.cursor -= cardsPerRow
			}

		case "down", "s", "j":
			// Calculate cards per row for navigation
			cardWidth := 26
			cardsPerRow := (m.windowWidth - 4) / cardWidth
			if cardsPerRow < 1 {
				cardsPerRow = 1
			}
			if m.cursor+cardsPerRow < len(m.interfaces) {
				m.cursor += cardsPerRow
			}

		case "left", "a", "h":
			if m.cursor > 0 {
				m.cursor--
			}

		case "right", "d", "l":
			if m.cursor < len(m.interfaces)-1 {
				m.cursor++
			}

		case "enter":
			if m.cursor >= len(m.interfaces) {
				return m, nil
			}

			m.selected = true
			m.selectedIface = m.interfaces[m.cursor]
			m.selectedAddrs = m.addrsByIface[m.selectedIface.Name]

			if len(m.selectedAddrs) == 0 {
				m.errorMsg = fmt.Sprintf("interface %s has no IP addresses", m.selectedIface.Name)
				m.selected = false
				return m, nil
			}

			var targetAddr string
			for _, addr := range m.selectedAddrs {
				if ipnet, ok := addr.(*net.IPNet); ok {
					if ipnet.IP.To4() != nil {
						targetAddr = ipnet.String()
						break
					}
				}
			}

			if targetAddr == "" {
				m.errorMsg = fmt.Sprintf("interface %s has no valid IPv4 addresses", m.selectedIface.Name)
				m.selected = false
				return m, nil
			}

			// Calculate total hosts for progress
			_, ipnet, _ := net.ParseCIDR(targetAddr)
			m.totalHosts = totalScanHosts(ipnet)

			m.scanning = true
			m.errorMsg = ""
			m.foundHosts = []string{}
			m.scannedCount = 0
			m.neighborSeen = 0
			m.neighborTotal = 0
			m.progressChan = make(chan scan.ScanProgress, 256)

			return m, performScan(m.scanner, m.selectedIface.Name, targetAddr, m.progressChan)
		}

	case scanProgressMsg:
		m.scannedCount = msg.Scanned
		if msg.Total > 0 {
			m.totalHosts = msg.Total
		}
		if msg.Phase == "neighbors" {
			m.neighborSeen = msg.PhaseScanned
			m.neighborTotal = msg.PhaseTotal
		}
		if msg.Phase == "sweep" {
			m.scannedCount = msg.PhaseScanned
			if msg.PhaseTotal > 0 {
				m.totalHosts = msg.PhaseTotal
			}
		}
		if msg.Host != "" {
			m.foundHosts = append(m.foundHosts, msg.Host)
		}
		return m, listenForProgress(m.progressChan)

	case scanCompleteMsg:
		m.scanning = false
		m.scanComplete = true
		return m, tea.Quit
	}

	return m, nil
}

func (m model) View() string {
	if m.scanning || m.scanComplete {
		var b strings.Builder

		// Header
		b.WriteString(titleStyle.Render(fmt.Sprintf("Scanning: %s", m.selectedIface.Name)))
		b.WriteString("\n")

		// Network info
		for _, addr := range m.selectedAddrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipnet.IP.To4() != nil {
					infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
					b.WriteString(infoStyle.Render(fmt.Sprintf("Network: %s", ipnet.String())) + "\n")
					break
				}
			}
		}

		// Show only sweep progress bar; neighbor discovery is summarized as text.
		sweepPercent := 0.0
		if m.totalHosts > 0 {
			sweepPercent = float64(m.scannedCount) / float64(m.totalHosts)
		}

		m.progress.Width = 50
		statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		b.WriteString(statsStyle.Render(fmt.Sprintf("Neighbor discovery %d/%d", m.neighborSeen, m.neighborTotal)) + "\n")
		b.WriteString(statsStyle.Render(fmt.Sprintf("Subnet sweep %d/%d", m.scannedCount, m.totalHosts)) + "\n")
		b.WriteString(m.progress.ViewAs(sweepPercent) + "\n")

		// Found hosts
		if len(m.foundHosts) > 0 {
			foundStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
			b.WriteString(foundStyle.Render(fmt.Sprintf("%d active:", len(m.foundHosts))) + "\n")

			hostStyle := lipgloss.NewStyle().Bold(true)
			portStyle := lipgloss.NewStyle()
			for _, host := range m.foundHosts {
				lines := strings.Split(host, "\n")
				b.WriteString(hostStyle.Render("• "+lines[0]) + "\n")
				for _, line := range lines[1:] {
					b.WriteString(portStyle.Render("    "+line) + "\n")
				}
			}
		} else if !m.scanComplete {
			emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("239")).Italic(true)
			b.WriteString(emptyStyle.Render("Searching...") + "\n")
		} else {
			emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("239")).Italic(true)
			b.WriteString(emptyStyle.Render("No hosts found") + "\n")
		}

		if m.scanning {
			helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			b.WriteString("\n" + helpStyle.Render("q: quit") + "\n")
		}

		return b.String()
	}

	// Interface selection view with cards
	var b strings.Builder

	// Title
	titleText := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Render("Nibble Network Scanner")
	b.WriteString(titleText + "\n")

	// Calculate cards per row based on terminal width
	cardWidth := 26 // 22 + 2 for border + 2 for spacing
	cardsPerRow := (m.windowWidth - 4) / cardWidth
	if cardsPerRow < 1 {
		cardsPerRow = 1
	}
	if m.windowWidth == 0 {
		cardsPerRow = 1 // Default before first resize
	}

	// Render cards in a grid
	var rows []string
	var currentRow []string

	for i, iface := range m.interfaces {
		isSelected := i == m.cursor
		style := cardStyle
		if isSelected {
			style = selectedCardStyle
		}

		// Card content
		var cardContent strings.Builder

		// Icon and name
		name := iface.Name
		icon := interfaceIcon(name)

		nameStyle := lipgloss.NewStyle().Bold(true)
		if isSelected {
			nameStyle = nameStyle.Foreground(lipgloss.Color("226"))
		}
		cardContent.WriteString(nameStyle.Render(icon+" "+name) + "\n")

		// IP addresses
		var addrs []string
		for _, addr := range m.addrsByIface[name] {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipnet.IP.To4() != nil {
					ones, _ := ipnet.Mask.Size()
					addrs = append(addrs, fmt.Sprintf("%s/%d", ipnet.IP.String(), ones))
				}
			}
		}

		addrStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		if len(addrs) > 0 {
			cardContent.WriteString(addrStyle.Render(addrs[0]))
		}

		card := style.Render(cardContent.String())
		currentRow = append(currentRow, card)

		// Start new row when we reach cardsPerRow or at the end
		if len(currentRow) == cardsPerRow || i == len(m.interfaces)-1 {
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
			currentRow = []string{}
		}
	}

	// Join all rows vertically
	b.WriteString(lipgloss.JoinVertical(lipgloss.Left, rows...))

	view := b.String()

	if m.errorMsg != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		view += "\n\n" + errorStyle.Render("⚠ Error: "+m.errorMsg)
	}

	// Add help text
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	view += "\n" + helpStyle.Render("arrows/wasd/hjkl: navigate • enter: select • ?: help • q: quit")

	// Show help overlay if requested
	if m.showHelp {
		helpBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("226")).
			Padding(0, 1).
			Width(56).
			Foreground(lipgloss.Color("15"))

		// Title with icon in upper-right corner
		helpTitle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true).
			Render("Nibble Network Scanner")

		iconStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true)

		// Create title row with icon on the right
		titleWidth := 54 // Box width minus padding
		titleText := helpTitle
		icon := iconStyle.Render("❓")
		spacer := strings.Repeat(" ", titleWidth-lipgloss.Width(titleText)-lipgloss.Width(icon))
		titleRow := titleText + spacer + icon

		helpContent := strings.Join([]string{
			titleRow,
			"Scans local networks for active hosts.",
			"• Checks TCP ports (SSH/HTTP/HTTPS/SMB/RDP)",
			"• Banner grabs services (SSH, HTTP Server)",
			"• Identifies hardware via MAC OUI (IEEE)",
			"• Runs 100 goroutines in parallel",
			"",
			"Press any key to close",
		}, "\n")

		helpOverlay := helpBox.Render(helpContent)

		// Center horizontally, align to top vertically with small margin
		return lipgloss.Place(
			lipgloss.Width(view),
			lipgloss.Height(view),
			lipgloss.Center,
			lipgloss.Top,
			helpOverlay,
			lipgloss.WithWhitespaceChars(" "),
		)
	}

	return docStyle.Render(view)
}
