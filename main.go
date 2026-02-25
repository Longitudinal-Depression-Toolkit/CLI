package main

import (
	"fmt"
	"os"

	app "ldt-toolkit-cli/internal"
)

func main() {
	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
