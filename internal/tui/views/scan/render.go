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

	if len(m.FoundHosts) > 0 && m.Results.Height > 0 {
		foundStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
		b.WriteString(foundStyle.Render(fmt.Sprintf("%d active:", len(m.FoundHosts))) + "\n")
		b.WriteString(m.Results.View() + "\n")
	} else if !m.ScanComplete {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("239")).Italic(true)
		b.WriteString(emptyStyle.Render("Searching...") + "\n")
	}
	if m.Scanning {
		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		b.WriteString("\n" + helpStyle.Render(renderHelpLine(maxWidth)) + "\n")
	}

	// Clear the rest of the screen when frame height shrinks so stale lines don't linger.
	return b.String() + "\x1b[J"
}
func FinalOutput(m Model) string {
	hosts := m.FoundHosts
	if len(m.FinalHosts) > 0 {
		hosts = m.FinalHosts
	}
	if len(hosts) == 0 {
		return "No hosts found"
	}

	foundStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
	return fmt.Sprintf("%s\n%s", foundStyle.Render(fmt.Sprintf("%d active:", len(hosts))), renderHostList(hosts))
}
