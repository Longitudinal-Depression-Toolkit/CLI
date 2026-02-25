package pivotlongtowide

import (
	"strings"

	"ldt-toolkit-cli/internal/screens/data_preprocessing/support/schema"
	"ldt-toolkit-cli/internal/screens/data_preprocessing/tools/spec"
	"ldt-toolkit-cli/internal/shared/components"
	"ldt-toolkit-cli/internal/shared/theme"
)

func Build(tool spec.Tool) spec.FlowSpec {
	flow := spec.Default(tool)
	flow.PromptTechniqueSelection = PromptTechniqueSelection
	flow.PromptTechniqueParams = PromptTechniqueParameters
	flow.SummaryPrinter = printSummary
	return flow
}

func printSummary(selected schema.Technique, result map[string]any) {
	components.PrintBlankLine()
	components.PrintLine(theme.App.SubtitleStyle().Render("Pivot long to wide complete"))
	components.PrintfLine("- Technique: %s", nonEmptyText(selected.Name, selected.ID))
	components.PrintfLine("- Output: %v", result["output_path"])
	components.PrintfLine("- Rows: %v", result["row_count"])
	components.PrintfLine("- Columns: %v", result["column_count"])
	components.PrintfLine("- Time column: %v", result["time_col"])
	components.PrintfLine("- Time label: %v", result["time_label"])
}

func nonEmptyText(primary string, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	if strings.TrimSpace(fallback) != "" {
		return fallback
	}
	return "Unnamed technique"
}
