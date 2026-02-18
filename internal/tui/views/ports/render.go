package portsview

import (
	"strings"

	"github.com/backendsystems/nibble/internal/tui/views/common"
	"github.com/charmbracelet/lipgloss"
)

func Render(m Model, maxWidth int) string {
	var b strings.Builder

	b.WriteString(common.TitleStyle.Render("Configure Scan Ports") + "\n")

	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	customStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	if m.PortPack == "default" {
		defaultStyle = defaultStyle.Foreground(lipgloss.Color("226")).Bold(true)
	} else {
		customStyle = customStyle.Foreground(lipgloss.Color("226")).Bold(true)
	}

	b.WriteString(defaultStyle.Render("default: 22,23,80,443,445,3389,8080") + "\n")
	customContent := m.CustomPorts
	if m.PortPack == "custom" {
		customContent = withCursor(m.CustomPorts, m.CustomCursor)
	}
	customLine := wrapCSVLineWithPrefix("custom:  ", customContent, maxWidth)
	invalidTokens := invalidPortTokens(m.ErrorMsg)
	if m.PortPack == "custom" && len(invalidTokens) > 0 {
		b.WriteString(highlightInvalidTokens(customLine, invalidTokens) + "\n")
	} else {
		b.WriteString(customStyle.Render(customLine) + "\n")
	}
	if m.PortPack == "custom" && strings.TrimSpace(m.CustomPorts) == "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true).Render("enter comma-separated ports, e.g. 22,80,443") + "\n")
	}

	if m.PortConfigLoc != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("saved at: "+m.PortConfigLoc) + "\n")
	}

	if m.ErrorMsg != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		b.WriteString("\n" + errorStyle.Render("Error: "+m.ErrorMsg) + "\n")
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	b.WriteString("\n" + helpStyle.Render(common.WrapWords(portsHelpText, maxWidth)))

	view := b.String()
	if m.ShowHelp {
		return renderHelpOverlay(view)
	}
	return common.DocStyle.Render(view)
}

func withCursor(s string, cursor int) string {
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(s) {
		cursor = len(s)
	}
	return s[:cursor] + "|" + s[cursor:]
}
