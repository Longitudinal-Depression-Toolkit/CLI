package benchstandardandlongitudinalml

import (
	"ldt-toolkit-cli/internal/screens/machine_learning/support/schema"
	"ldt-toolkit-cli/internal/screens/machine_learning/tools/prompting"
)

func PromptTechniqueSelection(title string, techniques []schema.Technique) (schema.Technique, error) {
	return prompting.PromptTechniqueSelection(title, techniques)
}

func PromptTechniqueParameters(technique schema.Technique) (map[string]any, error) {
	return prompting.PromptTechniqueParameters(technique)
}
