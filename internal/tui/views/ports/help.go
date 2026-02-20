package portsview

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderHelpOverlay(view string) string {
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
		"• type digits, commas, and ranges (e.g. 8000-9000)",
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
