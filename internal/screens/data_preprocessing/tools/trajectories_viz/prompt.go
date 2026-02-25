package trajectoriesviz

import (
	"ldt-toolkit-cli/internal/screens/data_preprocessing/support/schema"
	"ldt-toolkit-cli/internal/screens/data_preprocessing/tools/prompting"
)

func PromptTechniqueSelection(title string, techniques []schema.Technique) (schema.Technique, error) {
	return prompting.PromptTechniqueSelection(title, techniques)
}

func PromptTechniqueParameters(technique schema.Technique) (map[string]any, error) {
	return prompting.PromptTechniqueParameters(technique)
}
