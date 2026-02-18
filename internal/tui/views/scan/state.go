package scanview

import (
	"net"

	"github.com/backendsystems/nibble/internal/scanner"
	"github.com/charmbracelet/bubbles/progress"
)

type Model struct {
	NetworkScan   scanner.Scanner
	SelectedIface net.Interface
	SelectedAddrs []net.Addr
	Scanning      bool
	ScanComplete  bool
	FoundHosts    []string
	ScannedCount  int
	TotalHosts    int
	NeighborSeen  int
	NeighborTotal int
	ProgressChan  chan scanner.ProgressUpdate
	Progress      progress.Model
}
