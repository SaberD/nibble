package mainview

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
		Render("Nibble Network Scanner")

	iconStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Bold(true)

	titleWidth := 54
	icon := iconStyle.Render("❓")
	spacer := strings.Repeat(" ", titleWidth-lipgloss.Width(helpTitle)-lipgloss.Width(icon))
	titleRow := helpTitle + spacer + icon

	helpContent := strings.Join([]string{
		titleRow,
		"Scans local networks for active hosts.",
		"• Scans TCP ports",
		"  • Press p to configure ports",
		"• Grabs service banners (SSH, HTTP Server)",
		"• Identifies hardware via MAC OUI (IEEE)",
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
