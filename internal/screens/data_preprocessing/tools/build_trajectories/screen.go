package buildtrajectories

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
	components.PrintLine(theme.App.SubtitleStyle().Render("Build trajectories complete"))
	components.PrintfLine("- Technique: %s", nonEmptyText(selected.Name, selected.ID))
	components.PrintfLine("- Mode: %v", result["mode"])
	components.PrintfLine("- Builder: %v", result["builder"])
	components.PrintfLine("- Output: %v", result["output_path"])
	components.PrintfLine("- Rows: %v", result["row_count"])
	components.PrintfLine("- Columns: %v", result["column_count"])
	if n, ok := result["n_trajectories"]; ok && n != nil {
		components.PrintfLine("- Number of trajectories: %v", n)
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
