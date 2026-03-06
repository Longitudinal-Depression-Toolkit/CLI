package showtable

import (
	"strings"

	"ldt-toolkit-cli/internal/screens/data_preprocessing/support/schema"
	"ldt-toolkit-cli/internal/screens/data_preprocessing/tools/prompting"
)

func PromptTechniqueSelection(title string, techniques []schema.Technique) (schema.Technique, error) {
	return prompting.PromptTechniqueSelection(title, techniques)
}

func PromptTechniqueParameters(technique schema.Technique) (map[string]any, error) {
	technique = withBrowserOpenDefaults(technique)
	return prompting.PromptTechniqueParameters(technique)
}

func withBrowserOpenDefaults(technique schema.Technique) schema.Technique {
	parameters := make([]schema.Parameter, 0, len(technique.Parameters))
	for _, parameter := range technique.Parameters {
		if shouldDefaultToOpenInBrowser(parameter) {
			parameter.Default = true
		}
		parameters = append(parameters, parameter)
	}
	technique.Parameters = parameters
	return technique
}

func shouldDefaultToOpenInBrowser(parameter schema.Parameter) bool {
	typeName := strings.ToLower(strings.TrimSpace(parameter.Type))
	if typeName != "bool" && typeName != "boolean" {
		return false
	}

	text := strings.ToLower(strings.TrimSpace(strings.Join([]string{
		parameter.Key,
		parameter.Label,
		parameter.Hint,
	}, " ")))
	if text == "" || !strings.Contains(text, "browser") {
		return false
	}

	for _, token := range []string{"open", "show", "launch"} {
		if strings.Contains(text, token) {
			return true
		}
	}
	return false
}
