package scanview

import "github.com/backendsystems/nibble/internal/tui/views/common"

func renderHelpLine(maxWidth int) string {
	return common.WrapWords(scanHelpText, maxWidth)
}
