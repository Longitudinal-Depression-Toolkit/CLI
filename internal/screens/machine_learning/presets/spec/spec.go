package spec

import (
	"strings"

	"ldt-toolkit-cli/internal/screens/machine_learning/support/schema"
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
	return ModuleSpec{
		Title:   title,
		Summary: "Preset details are available for this route.",
		Lines: []string{
			"Use this preset route to review configuration and launch the workflow.",
		},
	}
}
