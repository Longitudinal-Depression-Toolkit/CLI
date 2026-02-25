package ui

import (
	"strings"

	"ldt-toolkit-cli/internal/screens/machine_learning/support/schema"
	"ldt-toolkit-cli/internal/shared/components"
)

func PrintTechniqueTable(techniques []schema.Technique) {
	rows := make([]components.TableRow, 0, len(techniques))
	for _, technique := range techniques {
		label := strings.TrimSpace(technique.Name)
		if label == "" {
			label = "Unnamed technique"
		}
		description := strings.TrimSpace(technique.Description)
		rows = append(rows, components.TableRow{Title: label, Description: description})
	}

	components.PrintBlankLine()
	components.PrintLine(components.RenderNameDescriptionTable(rows))
}
