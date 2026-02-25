package datapp

import (
	"fmt"
	"strings"
	"time"

	"ldt-toolkit-cli/internal/screens/data_preprocessing/support/ui"
	toolsmodule "ldt-toolkit-cli/internal/screens/data_preprocessing/tools"
	"ldt-toolkit-cli/internal/shared/components"
	"ldt-toolkit-cli/internal/shared/theme"
)

func runDataPreprocessingToolFlow(tool ToolConfig, listOnly bool) error {
	if !ToolIsRunnable(tool) {
		return renderUnavailableToolNotice(tool)
	}

	err := toolsmodule.Run(
		toolsmodule.Definition{
			ID:               tool.ID,
			Name:             tool.Name,
			Subtitle:         tool.Subtitle,
			TableTitle:       tool.TableTitle,
			SelectionTitle:   tool.SelectionTitle,
			CatalogOperation: tool.CatalogOperation,
			RunOperation:     tool.RunOperation,
		},
		listOnly,
		toolsmodule.Runtime{
			Execute:     executeBridge,
			InNavigator: currentNavigatorState,
		},
	)
	if toolsmodule.IsFlowCancelled(err) {
		return nil
	}
	return err
}

func renderUnavailableToolNotice(tool ToolConfig) error {
	title := strings.TrimSpace(tool.Name)
	if title == "" {
		title = "Data preprocessing tool"
	}

	summary := strings.TrimSpace(tool.Subtitle)
	if summary == "" {
		summary = strings.TrimSpace(tool.Description)
	}
	if summary == "" {
		summary = title
	}

	ui.PrepareActionScreen("Data Preprocessing Tools", summary, currentNavigatorState)
	components.PrintBlankLine()
	components.PrintLine(theme.App.SubtitleStyle().Render(title))
	components.PrintLine(theme.App.MutedTextStyle().Render(fmt.Sprintf("Status: %s", ToolStatusLabel(tool))))
	components.PrintLine(theme.App.MutedTextStyle().Render("This tool is incoming for the Go runtime."))
	return ui.RunExitCountdown("Returning to toolkit", 10*time.Second, currentNavigatorState)
}
