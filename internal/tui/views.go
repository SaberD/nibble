package tui

import (
	"fmt"
	"net"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	if m.scanning || m.scanComplete {
		return m.scanView()
	}
	return m.selectionView()
}

func (m model) scanView() string {
	var b strings.Builder

	// Header.
	b.WriteString(titleStyle.Render(fmt.Sprintf("Scanning: %s", m.selectedIface.Name)))
	b.WriteString("\n")

	// Network info.
	for _, addr := range m.selectedAddrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			b.WriteString(infoStyle.Render(fmt.Sprintf("Network: %s", ipnet.String())) + "\n")
			break
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

	// Found hosts.
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

func (m model) selectionView() string {
	var b strings.Builder

	titleText := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Render("Nibble Network Scanner")
	b.WriteString(titleText + "\n")

	cardsPerRow := m.cardsPerRow()
	if m.windowWidth == 0 {
		cardsPerRow = 1 // Default before first resize.
	}

	// Render cards in a grid.
	var rows []string
	var currentRow []string

	for i, iface := range m.interfaces {
		card := m.renderInterfaceCard(i, iface)
		currentRow = append(currentRow, card)

		if len(currentRow) == cardsPerRow || i == len(m.interfaces)-1 {
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
			currentRow = nil
		}
	}

	b.WriteString(lipgloss.JoinVertical(lipgloss.Left, rows...))
	view := b.String()

	if m.errorMsg != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		view += "\n\n" + errorStyle.Render("⚠ Error: "+m.errorMsg)
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	view += "\n" + helpStyle.Render("arrows/wasd/hjkl: navigate • enter: select • ?: help • q: quit")

	if m.showHelp {
		return m.renderHelpOverlay(view)
	}

	return docStyle.Render(view)
}

func (m model) renderInterfaceCard(index int, iface net.Interface) string {
	isSelected := index == m.cursor
	style := cardStyle
	if isSelected {
		style = selectedCardStyle
	}

	var cardContent strings.Builder
	name := iface.Name
	icon := interfaceIcon(name)

	nameStyle := lipgloss.NewStyle().Bold(true)
	if isSelected {
		nameStyle = nameStyle.Foreground(lipgloss.Color("226"))
	}
	cardContent.WriteString(nameStyle.Render(icon+" "+name) + "\n")

	addrs := m.interfaceIPv4Labels(name)
	addrStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	if len(addrs) > 0 {
		cardContent.WriteString(addrStyle.Render(addrs[0]))
	}

	return style.Render(cardContent.String())
}

func (m model) interfaceIPv4Labels(name string) []string {
	var labels []string
	for _, addr := range m.addrsByIface[name] {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			ones, _ := ipnet.Mask.Size()
			labels = append(labels, fmt.Sprintf("%s/%d", ipnet.IP.String(), ones))
		}
	}
	return labels
}

func (m model) renderHelpOverlay(view string) string {
	helpBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("226")).
		Padding(0, 1).
		Width(56).
		Foreground(lipgloss.Color("15"))

	helpTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Bold(true).
		Render("Nibble Network Scanner")

	iconStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Bold(true)

	titleWidth := 54 // Box width minus padding.
	icon := iconStyle.Render("❓")
	spacer := strings.Repeat(" ", titleWidth-lipgloss.Width(helpTitle)-lipgloss.Width(icon))
	titleRow := helpTitle + spacer + icon

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

	return lipgloss.Place(
		lipgloss.Width(view),
		lipgloss.Height(view),
		lipgloss.Center,
		lipgloss.Top,
		helpOverlay,
		lipgloss.WithWhitespaceChars(" "),
	)
}
