package tools

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"ldt-toolkit-cli/internal/screens/data_preprocessing/support/schema"
	"ldt-toolkit-cli/internal/screens/data_preprocessing/support/ui"
	"ldt-toolkit-cli/internal/screens/data_preprocessing/tools/prompting"
	"ldt-toolkit-cli/internal/shared/components"
	"ldt-toolkit-cli/internal/shared/theme"
)

func Run(def Definition, listOnly bool, runtime Runtime) error {
	if runtime.Execute == nil {
		return errors.New("data preprocessing tools runtime has no bridge executor")
	}

	screenTool := Tool{
		ID:             def.ID,
		Name:           def.Name,
		Subtitle:       def.Subtitle,
		TableTitle:     def.TableTitle,
		SelectionTitle: def.SelectionTitle,
	}
	spec, ok := Resolve(screenTool)
	if !ok {
		spec = Default(screenTool)
	}

	ui.PrepareActionScreen(spec.Title, spec.Subtitle, runtime.InNavigator)

	techniques, err := fetchToolCatalog(def, runtime.Execute)
	if err != nil {
		return err
	}
	if len(techniques) == 0 {
		return fmt.Errorf("no techniques were found for %s", strings.ToLower(spec.Title))
	}

	var selected schema.Technique
	if len(techniques) == 1 {
		selected = techniques[0]
		if listOnly {
			return nil
		}
	} else {
		ui.PrintTechniqueTable(techniques)
		if listOnly {
			return nil
		}

		selectionTitle := strings.TrimSpace(spec.SelectionTitle)
		if selectionTitle == "" {
			selectionTitle = fmt.Sprintf("%s technique", strings.TrimSpace(spec.Title))
		}

		if spec.PromptTechniqueSelection == nil {
			spec.PromptTechniqueSelection = prompting.PromptTechniqueSelection
		}
		selected, err = spec.PromptTechniqueSelection(selectionTitle, techniques)
		if err != nil {
			return err
		}
	}

	if spec.PromptTechniqueParams == nil {
		spec.PromptTechniqueParams = prompting.PromptTechniqueParameters
	}
	params, err := spec.PromptTechniqueParams(selected)
	if err != nil {
		return err
	}

	result, err := runtime.Execute(
		def.RunOperation,
		map[string]any{
			"technique": selected.ID,
			"params":    params,
		},
	)
	if err != nil {
		return err
	}

	if spec.SummaryPrinter != nil {
		spec.SummaryPrinter(selected, result)
	} else {
		printGenericSummary(selected, result)
	}

	return ui.RunExitCountdown("Returning to toolkit", 10*time.Second, runtime.InNavigator)
}

func fetchToolCatalog(def Definition, execute BridgeExecutor) ([]schema.Technique, error) {
	result, err := execute(def.CatalogOperation, map[string]any{})
	if err != nil {
		return nil, err
	}
	return schema.DecodeTechniquePayload(result)
}

func printGenericSummary(selected schema.Technique, result map[string]any) {
	components.PrintBlankLine()
	components.PrintLine(theme.App.SubtitleStyle().Render("Run complete"))
	components.PrintfLine("- Technique: %s", nonEmptyText(selected.Name, selected.ID))
	for key, value := range result {
		components.PrintfLine("- %s: %v", strings.TrimSpace(key), value)
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
