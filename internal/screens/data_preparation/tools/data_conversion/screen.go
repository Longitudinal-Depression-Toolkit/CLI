package dataconversion

import (
	"fmt"
	"strings"

	"ldt-toolkit-cli/internal/screens/data_preparation/support/schema"
	"ldt-toolkit-cli/internal/screens/data_preparation/tools/spec"
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
	components.PrintLine(theme.App.SubtitleStyle().Render("Data conversion complete"))
	components.PrintfLine("- Technique: %s", nonEmptyText(selected.Name, selected.ID))
	components.PrintfLine("- Mode: %v", result["mode"])

	if strings.EqualFold(fmt.Sprintf("%v", result["mode"]), "single") {
		components.PrintfLine("- Input: %v", result["input_path"])
		components.PrintfLine("- Output: %v", result["output_path"])
		components.PrintfLine("- Rows: %v", result["row_count"])
		components.PrintfLine("- Columns: %v", result["column_count"])
		return
	}

	components.PrintfLine("- Input folder: %v", result["input_folder"])
	components.PrintfLine("- Output folder: %v", result["output_folder"])
	components.PrintfLine("- Total files: %v", result["total_files"])
	components.PrintfLine("- Converted files: %v", result["converted_files"])
	components.PrintfLine("- Failed files: %v", result["failed_files"])
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
