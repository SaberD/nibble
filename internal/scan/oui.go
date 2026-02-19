package scan

import (
	"encoding/csv"
	"io"
	"strings"

	_ "embed"
)

//go:embed oui.csv
var ouiCsv string

// ouiMap maps MAC prefixes (lowercase, colon-separated) to manufacturer names.
// Loaded from the embedded IEEE OUI database at init time.
var ouiMap map[string]string

func init() {
	ouiMap = make(map[string]string, 40000)
	r := csv.NewReader(strings.NewReader(ouiCsv))
	firstRow := true
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return
		}
		if firstRow {
			firstRow = false
			continue // skip header
		}
		if len(rec) < 3 {
			continue
		}
		// Assignment is 6 hex chars like "286FB9", convert to "28:6f:b9".
		hex := strings.ToLower(strings.TrimSpace(rec[1]))
		if len(hex) != 6 {
			continue
		}
		prefix := hex[0:2] + ":" + hex[2:4] + ":" + hex[4:6]
		ouiMap[prefix] = strings.TrimSpace(rec[2])
	}
}

// VendorFromMac returns the hardware manufacturer from a MAC address OUI prefix.
func VendorFromMac(mac string) string {
	mac = strings.ToLower(mac)
	// the first 3 bytes in a MAC "xx:xx:xx" maps to a vendor.
	if len(mac) >= 8 {
		if vendor, ok := ouiMap[mac[:8]]; ok {
			return vendor
		}
	}
	if mac != "" {
		return strings.ToUpper(mac)
	}
	return ""
}
