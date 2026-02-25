package components

import (
	"fmt"
)

func clearTerminalScreen() {
	_, _ = fmt.Print("\033[2J\033[H")

	header := RenderScreenHeader(0)
	if header == "" {
		return
	}

	_, _ = fmt.Print(header)
	_, _ = fmt.Print("\n\n")
}
