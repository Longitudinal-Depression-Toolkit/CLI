package spec

import (
	"strings"

	"ldt-toolkit-cli/internal/screens/data_preparation/support/schema"
)

type Tool struct {
	ID             string
	Name           string
	Subtitle       string
	TableTitle     string
	SelectionTitle string
}

type PromptTechniqueSelectionFunc func(title string, techniques []schema.Technique) (schema.Technique, error)
type PromptTechniqueParametersFunc func(technique schema.Technique) (map[string]any, error)

type FlowSpec struct {
	Title                    string
	Subtitle                 string
	TableTitle               string
	SelectionTitle           string
	PromptTechniqueSelection PromptTechniqueSelectionFunc
	PromptTechniqueParams    PromptTechniqueParametersFunc
	SummaryPrinter           func(selected schema.Technique, result map[string]any)
}

func Default(tool Tool) FlowSpec {
	return FlowSpec{
		Title:          NonEmpty(tool.Name, "Unnamed tool"),
		Subtitle:       tool.Subtitle,
		TableTitle:     tool.TableTitle,
		SelectionTitle: tool.SelectionTitle,
		SummaryPrinter: nil,
	}
}

func NonEmpty(primary string, fallback string) string {
	trimmed := strings.TrimSpace(primary)
	if trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(fallback)
}
