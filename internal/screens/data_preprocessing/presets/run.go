package presets

import (
	"fmt"
	"strings"
	"time"

	"ldt-toolkit-cli/internal/screens/data_preprocessing/support/ui"
	"ldt-toolkit-cli/internal/shared/components"
	"ldt-toolkit-cli/internal/shared/theme"
)

func Run(preset Preset, runtime Runtime) error {
	status := StatusLabel(preset)
	title := strings.TrimSpace(preset.Name)
	if title == "" {
		title = "Data preprocessing preset"
	}

	summary := strings.TrimSpace(preset.Description)
	if summary == "" {
		summary = title
	}

	ui.PrepareActionScreen("Data Preprocessing Presets", summary, runtime.InNavigator)
	components.PrintBlankLine()
	components.PrintLine(theme.App.SubtitleStyle().Render(title))

	if IsRunnable(preset) {
		components.PrintLine(theme.App.MutedTextStyle().Render("This preset is available but not wired yet in the Go runtime."))
	} else {
		components.PrintLine(theme.App.MutedTextStyle().Render(fmt.Sprintf("Preset status: %s", status)))
		components.PrintLine(theme.App.MutedTextStyle().Render("This preset is incoming. Use tools while this workflow is being integrated."))
	}

	return ui.RunExitCountdown("Returning to toolkit", 10*time.Second, runtime.InNavigator)
}
