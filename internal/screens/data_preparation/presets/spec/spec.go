package spec

import (
	"strings"

	"ldt-toolkit-cli/internal/screens/data_preparation/support/schema"
)

type Preset = schema.Preset

type ModuleSpec struct {
	Title   string
	Summary string
	Lines   []string
}

func Default(preset Preset) ModuleSpec {
	title := strings.TrimSpace(preset.Name)
	if title == "" {
		title = "Unnamed preset"
	}
	summary := "Preset details are available for this route."
	return ModuleSpec{
		Title:   title,
		Summary: summary,
		Lines: []string{
			"Use this preset route to review configuration and launch the workflow.",
		},
	}
}
