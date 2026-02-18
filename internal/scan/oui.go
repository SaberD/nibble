package scan

import (
	"encoding/csv"
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
	records, err := r.ReadAll()
	if err != nil {
		return
	}
	for _, rec := range records[1:] { // skip header
		if len(rec) < 3 {
			continue
		}
		// Assignment is 6 hex chars like "286FB9", convert to "28:6f:b9".
		hex := strings.ToLower(rec[1])
		if len(hex) != 6 {
			continue
		}
		prefix := hex[0:2] + ":" + hex[2:4] + ":" + hex[4:6]
		ouiMap[prefix] = rec[2]
	}
}

// vendorFromMac returns the hardware manufacturer from a MAC address OUI prefix.
func vendorFromMac(mac string) string {
	mac = strings.ToLower(mac)
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
