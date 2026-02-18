package tui

import (
	"net"
	"os"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scan"
	"github.com/backendsystems/nibble/internal/scanner"
	mainview "github.com/backendsystems/nibble/internal/tui/views/main"
	portsview "github.com/backendsystems/nibble/internal/tui/views/ports"
	scanview "github.com/backendsystems/nibble/internal/tui/views/scan"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
)

type activeView int

const (
	viewMain activeView = iota
	viewPorts
	viewScan
)

type model struct {
	active      activeView
	windowW     int
	cardsPerRow int
	main        mainview.Model
	ports       portsview.Model
	scan        scanview.Model
}

func Run(networkScanner scanner.Scanner, ifaces []net.Interface, addrsByIface map[string][]net.Addr) error {
	cfg, _ := ports.LoadConfig()
	pack := cfg.Mode
	if pack == "" || !ports.IsValidPack(pack) {
		pack = "default"
	}
	if netScanner, ok := networkScanner.(*scan.NetScanner); ok {
		if resolvedPorts, err := ports.Resolve(pack, cfg.Custom, ""); err == nil {
			netScanner.Ports = resolvedPorts
		}
	}

	initialWindowW, initialCardsPerRow := initialLayoutMetrics()

	initialModel := model{
		active:      viewMain,
		windowW:     initialWindowW,
		cardsPerRow: initialCardsPerRow,
		main: mainview.Model{
			Interfaces:   ifaces,
			InterfaceMap: addrsByIface,
		},
		ports: portsview.Model{
			PortPack:    pack,
			CustomPorts: cfg.Custom,
			NetworkScan: networkScanner,
		},
		scan: scanview.Model{
			NetworkScan: networkScanner,
			Progress: progress.New(
				progress.WithScaledGradient("#FFD700", "#B8B000"),
			),
		},
	}

	prog := tea.NewProgram(initialModel)
	_, err := prog.Run()
	return err
}

func (m model) Init() tea.Cmd {
	if m.ports.PortPack == "" {
		m.ports.PortPack = "default"
	}
	if m.ports.PortConfigLoc == "" {
		if path, err := ports.ConfigPath(); err == nil {
			m.ports.PortConfigLoc = path
		}
	}
	if m.ports.CustomCursor < 0 || m.ports.CustomCursor > len(m.ports.CustomPorts) {
		m.ports.CustomCursor = len(m.ports.CustomPorts)
	}
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		return m, nil
	}

	switch m.active {
	case viewScan:
		result := m.scan.Update(msg)
		if !result.Handled {
			return m, nil
		}
		m.scan = result.Model
		if result.Quit {
			return m, tea.Quit
		}
		return m, result.Cmd
	case viewPorts:
		key, ok := msg.(tea.KeyMsg)
		if !ok {
			return m, nil
		}
		result := m.ports.Update(key)
		m.ports = result.Model
		if result.Quit {
			return m, tea.Quit
		}
		if result.Done {
			m.main.ErrorMsg = ""
			m.active = viewMain
		}
		return m, nil
	case viewMain:
		key, ok := msg.(tea.KeyMsg)
		if !ok {
			return m, nil
		}
		result := m.main.Update(key)
		m.main = result.Model
		if result.Quit {
			return m, tea.Quit
		}
		if result.OpenPorts {
			m.ports.ShowHelp = false
			m.ports.CustomCursor = len(m.ports.CustomPorts)
			m.active = viewPorts
			return m, nil
		}
		if result.StartScan {
			m.main.ErrorMsg = ""
			nextScan, cmd := m.scan.Start(
				result.Selection.Iface,
				result.Selection.Addrs,
				result.Selection.TotalHosts,
				result.Selection.TargetAddr,
			)
			m.scan = nextScan
			m.active = viewScan
			return m, cmd
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m model) View() string {
	maxWidth := 72
	if m.windowW > 8 {
		maxWidth = m.windowW - 4
	}
	cardsPerRow := m.cardsPerRow
	if cardsPerRow == 0 {
		cardsPerRow = 1
	}

	switch m.active {
	case viewScan:
		return scanview.Render(m.scan, maxWidth)
	case viewPorts:
		return portsview.Render(m.ports, maxWidth)
	default:
		return mainview.Render(m.main, maxWidth, cardsPerRow)
	}
}

func initialLayoutMetrics() (windowW int, cardsPerRow int) {
	cardsPerRow = 1
	fd := os.Stdout.Fd()
	if !term.IsTerminal(fd) {
		return 0, cardsPerRow
	}

	width, _, err := term.GetSize(fd)
	if err != nil || width <= 0 {
		return 0, cardsPerRow
	}

	return width, mainview.CardsPerRow(width)
}
