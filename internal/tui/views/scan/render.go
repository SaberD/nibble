package scanview

import (
	"fmt"
	"net"
	"strings"

	"github.com/backendsystems/nibble/internal/tui/views/common"
	"github.com/charmbracelet/lipgloss"
)

func Render(m Model, maxWidth int) string {
	var b strings.Builder

	b.WriteString(common.TitleStyle.Render(fmt.Sprintf("Scanning: %s", m.SelectedIface.Name)))
	b.WriteString("\n")

	for _, addr := range m.SelectedAddrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			b.WriteString(infoStyle.Render(fmt.Sprintf("Network: %s", ipnet.String())) + "\n")
			break
		}
	}

	statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	b.WriteString(statsStyle.Render(fmt.Sprintf("Neighbor discovery %d/%d", m.NeighborSeen, m.NeighborTotal)) + "\n")
	b.WriteString(statsStyle.Render(fmt.Sprintf("Subnet sweep %d/%d", m.ScannedCount, m.TotalHosts)) + "\n")

	sweepPercent := 0.0
	if m.TotalHosts > 0 {
		sweepPercent = float64(m.ScannedCount) / float64(m.TotalHosts)
	}
	progressModel := m.Progress
	progressModel.Width = 50
	b.WriteString(progressModel.ViewAs(sweepPercent) + "\n")

	if len(m.FoundHosts) > 0 {
		foundStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
		b.WriteString(foundStyle.Render(fmt.Sprintf("%d active:", len(m.FoundHosts))) + "\n")

		hostStyle := lipgloss.NewStyle().Bold(true)
		portStyle := lipgloss.NewStyle()
		for _, host := range m.FoundHosts {
			lines := strings.Split(host, "\n")
			b.WriteString(hostStyle.Render("• "+lines[0]) + "\n")
			for _, line := range lines[1:] {
				b.WriteString(portStyle.Render("    "+line) + "\n")
			}
		}
	} else if !m.ScanComplete {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("239")).Italic(true)
		b.WriteString(emptyStyle.Render("Searching...") + "\n")
	} else {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("239")).Italic(true)
		b.WriteString(emptyStyle.Render("No hosts found") + "\n")
	}

	if m.Scanning {
		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		b.WriteString("\n" + helpStyle.Render(renderHelpLine(maxWidth)) + "\n")
	}

	return b.String()
}
