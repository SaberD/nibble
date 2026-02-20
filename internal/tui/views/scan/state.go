package scanview

import (
	"net"

	"github.com/backendsystems/nibble/internal/scanner"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
)

type Model struct {
	NetworkScan      scanner.Scanner
	SelectedIface    net.Interface
	SelectedAddrs    []net.Addr
	Scanning         bool
	ScanComplete     bool
	ShouldPrintFinal bool
	FoundHosts       []string
	FinalHosts       []string
	ScannedCount     int
	TotalHosts       int
	NeighborSeen     int
	NeighborTotal    int
	ProgressChan     chan scanner.ProgressUpdate
	Progress         progress.Model
	Results          viewport.Model
}
