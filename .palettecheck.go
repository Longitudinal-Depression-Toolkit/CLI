package main

import (
	"fmt"
	cp "ldt-toolkit-cli/internal/command_palette"
	"strings"
)

func dump(width int) {
	m, err := cp.New(nil)
	if err != nil {
		panic(err)
	}
	m.Resize(width)
	_ = m.Open()
	view := m.View()
	lines := strings.Split(view, "\n")
	fmt.Printf("\n--- width=%d lines=%d ---\n", width, len(lines))
	for i, ln := range lines {
		r := []rune(ln)
		last := ""
		if len(r) > 0 {
			last = string(r[len(r)-1])
		}
		fmt.Printf("%03d len=%3d last=%q\n", i+1, len(r), last)
	}
	fmt.Println(lines[0])
	fmt.Println(lines[len(lines)-1])
}

func main() {
	dump(120)
	dump(160)
	dump(200)
}
