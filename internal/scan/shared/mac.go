package shared

import (
	"strconv"
	"strings"
)

func NormalizeMAC(mac string) string {
	mac = strings.ToLower(strings.TrimSpace(strings.ReplaceAll(mac, "-", ":")))
	parts := strings.Split(mac, ":")
	if len(parts) != 6 {
		return ""
	}

	for i, part := range parts {
		if len(part) == 0 || len(part) > 2 {
			return ""
		}
		value, err := strconv.ParseUint(part, 16, 8)
		if err != nil {
			return ""
		}
		parts[i] = strings.ToLower(strconv.FormatUint(value, 16))
		if len(parts[i]) == 1 {
			parts[i] = "0" + parts[i]
		}
	}

	return strings.Join(parts, ":")
}
