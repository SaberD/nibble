package main

import (
	"fmt"
	"net"
	"strings"

	"nibble/internal/scan"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	itemStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
)

type ifaceInfo struct {
	iface net.Interface
	addrs []net.Addr
}

type model struct {
	ifaces        []ifaceInfo // Changed to hold interface and its addresses
	cursor        int
	selected      bool
	selectedIface ifaceInfo // Changed to hold selected interface and its addresses
}

func (m model) Init() tea.Cmd {
	return nil // no initial command
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			if m.cursor < len(m.ifaces)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "enter":
			m.selected = true
			m.selectedIface = m.ifaces[m.cursor]
			results := scan.PerformScan(m.selectedIface.addrs[0].String())
			prettyPrintResults(results)
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	var b strings.Builder
	b.WriteString("Choose an interface:\n\n")

	for i, ifaceInfo := range m.ifaces {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Show interface name
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, ifaceInfo.iface.Name))
		// Show IP addresses for the interface
		for _, addr := range ifaceInfo.addrs {
			b.WriteString(fmt.Sprintf("       %s\n", addr.String()))
		}
	}

	b.WriteString("\nPress q to quit.\n")

	return b.String()
}

func prettyPrintResults(results []string) {
	title := titleStyle.Render("Scan Results:")
	fmt.Println(title)

	for _, result := range results {
		item := itemStyle.Render(result)
		fmt.Println(" •", item)
	}
}
