package removecolumns

import (
	"fmt"
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
	components.PrintLine(theme.App.SubtitleStyle().Render("Remove columns complete"))
	components.PrintfLine("- Technique: %s", nonEmptyText(selected.Name, selected.ID))
	components.PrintfLine("- Output: %v", result["output_path"])
	components.PrintfLine("- Rows: %v", result["row_count"])
	components.PrintfLine("- Columns: %v", result["column_count"])
	if removed, ok := result["removed_columns"].([]any); ok && len(removed) > 0 {
		values := make([]string, 0, len(removed))
		for _, value := range removed {
			values = append(values, fmt.Sprintf("%v", value))
		}
		components.PrintfLine("- Removed: %s", strings.Join(values, ", "))
	}
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
