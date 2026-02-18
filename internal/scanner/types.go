package scanner

type ProgressUpdate interface {
	isProgressUpdate()
}

// NeighborProgress represents progress during the neighbor discovery phase.
type NeighborProgress struct {
	Host       string // Optional host line found during neighbor discovery.
	TotalHosts int    // Overall total hosts in the subnet sweep.
	Seen       int    // Neighbors processed so far.
	Total      int    // Total neighbors to process.
}

func (NeighborProgress) isProgressUpdate() {}

// SweepProgress represents progress during the subnet sweep phase.
type SweepProgress struct {
	Host       string // Optional host line found during sweep.
	TotalHosts int    // Overall total hosts in the subnet sweep.
	Scanned    int    // Hosts scanned so far in sweep.
	Total      int    // Total hosts in sweep phase.
}

func (SweepProgress) isProgressUpdate() {}

// Scanner abstracts network scanning so real and demo modes share the same code path.
type Scanner interface {
	ScanNetwork(ifaceName, subnet string, progressChan chan<- ProgressUpdate)
}
