package components

import (
	"fmt"
	"strings"
)

func PrintBlankLine() {
	fmt.Println()
}

func PrintLine(text string) {
	fmt.Println(ApplyLeftLayoutMargin(text))
}

func PrintfLine(format string, args ...any) {
	PrintLine(fmt.Sprintf(format, args...))
}

func PrintBlock(text string) {
	trimmedRight := strings.TrimRight(text, "\n")
	if trimmedRight == "" {
		PrintBlankLine()
		return
	}
	PrintLine(trimmedRight)
}
