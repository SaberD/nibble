package scanview

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

const (
	defaultResultsWidth  = 72
	defaultResultsHeight = 10
	minResultsHeight     = 3
)

func (m Model) SetViewportSize(maxWidth, windowHeight int) Model {
	width := maxWidth
	if width <= 0 {
		width = defaultResultsWidth
	}

	height := defaultResultsHeight
	if windowHeight > 0 {
		// Header + status + progress + footer reserve space so results stay bounded.
		reserved := 10
		if m.Scanning {
			reserved = 12
		}
		height = windowHeight - reserved
	}
	if height < minResultsHeight {
		height = minResultsHeight
	}

	if m.Results.Width == 0 || m.Results.Height == 0 {
		m.Results = viewport.New(width, height)
	} else {
		m.Results.Width = width
		m.Results.Height = height
	}

	if m.Results.PastBottom() {
		m.Results.GotoBottom()
	}
	return m
}

func (m Model) RefreshResults(stickToBottom bool) Model {
	atBottom := m.Results.AtBottom()
	m.Results.SetContent(renderHostList(m.FoundHosts))
	if stickToBottom && atBottom {
		m.Results.GotoBottom()
	}
	return m
}

func renderHostList(hosts []string) string {
	hostStyle := lipgloss.NewStyle().Bold(true)
	portStyle := lipgloss.NewStyle()

	var b strings.Builder
	for i, host := range hosts {
		lines := strings.Split(host, "\n")
		b.WriteString(hostStyle.Render("• " + lines[0]))
		if i < len(hosts)-1 || len(lines) > 1 {
			b.WriteString("\n")
		}
		for j, line := range lines[1:] {
			b.WriteString(portStyle.Render("    " + line))
			if i < len(hosts)-1 || j < len(lines[1:])-1 {
				b.WriteString("\n")
			}
		}
	}
	return b.String()
}
