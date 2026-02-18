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
	if m.editingPorts {
		return m.portsView()
	}
	return m.selectionView()
}

func (m model) portsView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Configure Scan Ports") + "\n")

	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	customStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	if m.portPack == "default" {
		defaultStyle = defaultStyle.Foreground(lipgloss.Color("226")).Bold(true)
	} else {
		customStyle = customStyle.Foreground(lipgloss.Color("226")).Bold(true)
	}

	b.WriteString(defaultStyle.Render("default: 22,23,80,443,445,3389,8080") + "\n")
	customContent := m.customPorts
	if m.portPack == "custom" {
		customContent = m.customPortsWithCursor()
	}
	maxWidth := 72
	if m.windowWidth > 8 {
		maxWidth = m.windowWidth - 4
	}
	customLine := wrapCSVLineWithPrefix("custom:  ", customContent, maxWidth)
	invalidTokens := invalidPortTokens(m.errorMsg)
	if m.portPack == "custom" && len(invalidTokens) > 0 {
		b.WriteString(highlightInvalidTokens(customLine, invalidTokens) + "\n")
	} else {
		b.WriteString(customStyle.Render(customLine) + "\n")
	}
	if m.portPack == "custom" && strings.TrimSpace(m.customPorts) == "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true).Render("enter comma-separated ports, e.g. 22,80,443") + "\n")
	}

	if m.portConfigLoc != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("saved at: "+m.portConfigLoc) + "\n")
	}

	if m.errorMsg != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		b.WriteString("\n" + errorStyle.Render("Error: "+m.errorMsg) + "\n")
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	helpText := "tab • ←/→ a/d h/l • type • backspace: remove • delete: clear all • enter • ?: help • q: quit"
	b.WriteString("\n" + helpStyle.Render(wrapWords(helpText, maxWidth)))

	view := b.String()
	if m.showHelp {
		return m.renderPortsHelpOverlay(view)
	}
	return docStyle.Render(view)
}

func (m model) customPortsWithCursor() string {
	cursor := m.customCursor
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(m.customPorts) {
		cursor = len(m.customPorts)
	}
	return m.customPorts[:cursor] + "|" + m.customPorts[cursor:]
}

func wrapCSVLineWithPrefix(prefix, content string, maxWidth int) string {
	if content == "" {
		return prefix
	}
	if maxWidth <= len(prefix)+1 {
		return prefix + content
	}

	indent := strings.Repeat(" ", len(prefix))
	tokens := strings.Split(content, ",")
	lines := make([]string, 0, 4)
	current := prefix

	for i, token := range tokens {
		segment := token
		if i < len(tokens)-1 {
			segment += ","
		}

		// Break on comma boundaries when the current line is full.
		if len(current)+len(segment) > maxWidth && len(current) > len(prefix) {
			lines = append(lines, current)
			current = indent + segment
			continue
		}
		current += segment
	}

	lines = append(lines, current)
	return strings.Join(lines, "\n")
}

func wrapWords(s string, maxWidth int) string {
	if maxWidth <= 0 || len(s) <= maxWidth {
		return s
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}

	lines := []string{words[0]}
	for _, w := range words[1:] {
		last := lines[len(lines)-1]
		if len(last)+1+len(w) <= maxWidth {
			lines[len(lines)-1] = last + " " + w
			continue
		}
		lines = append(lines, w)
	}
	return strings.Join(lines, "\n")
}

func invalidPortTokens(errMsg string) []string {
	const prefix = "invalid ports: "
	lower := strings.ToLower(errMsg)
	if !strings.Contains(lower, prefix) {
		return nil
	}
	i := strings.Index(lower, prefix)
	if i < 0 {
		return nil
	}
	rest := strings.TrimSpace(errMsg[i+len(prefix):])
	if rest == "" {
		return nil
	}
	parts := strings.Split(rest, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s == "" {
			continue
		}
		out = append(out, s)
	}
	return out
}

func highlightInvalidTokens(s string, tokens []string) string {
	invalidStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	for _, token := range tokens {
		if token == "" {
			continue
		}
		start := 0
		for {
			idx := strings.Index(s[start:], token)
			if idx < 0 {
				break
			}
			idx += start
			end := idx + len(token)

			prevOK := idx == 0 || strings.ContainsRune(" ,:|\n", rune(s[idx-1]))
			nextOK := end == len(s) || strings.ContainsRune(",|\n ", rune(s[end]))
			if prevOK && nextOK {
				s = s[:idx] + invalidStyle.Render(token) + s[end:]
				break
			}
			start = end
		}
	}
	return s
}

func (m model) scanView() string {
	var b strings.Builder
	maxWidth := 72
	if m.windowWidth > 8 {
		maxWidth = m.windowWidth - 4
	}

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
		b.WriteString("\n" + helpStyle.Render(wrapWords("q: quit", maxWidth)) + "\n")
	}

	return b.String()
}

func (m model) selectionView() string {
	var b strings.Builder
	maxWidth := 72
	if m.windowWidth > 8 {
		maxWidth = m.windowWidth - 4
	}

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
		view += "\n\n" + errorStyle.Render("Error: "+m.errorMsg)
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	helpText := "←/→ a/d h/l • p: ports • ?: help • q: quit"
	view += "\n" + helpStyle.Render(wrapWords(helpText, maxWidth))

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
		"• Checks configurable TCP ports",
		"• Banner grabs services (SSH, HTTP Server)",
		"• Identifies hardware via MAC OUI (IEEE)",
		"• Runs 100 goroutines in parallel",
		"• Press p to configure ports (default/custom)",
		"",
		"any key: close",
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

func (m model) renderPortsHelpOverlay(view string) string {
	helpBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("226")).
		Padding(0, 1).
		Width(56).
		Foreground(lipgloss.Color("15"))

	helpTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Bold(true).
		Render("Port Configuration")

	iconStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Bold(true)

	titleWidth := 54
	icon := iconStyle.Render("❓")
	spacer := strings.Repeat(" ", titleWidth-lipgloss.Width(helpTitle)-lipgloss.Width(icon))
	titleRow := helpTitle + spacer + icon

	helpContent := strings.Join([]string{
		titleRow,
		"Configure which ports get scanned.",
		"• tab: switch default/custom mode",
		"• ←/→ or a/d or h/l: move cursor in custom list",
		"• type digits and commas to edit custom ports",
		"• backspace: remove",
		"• delete: clear all",
		"• q: quit",
		"• enter: save and return",
		"",
		"any key: close",
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
