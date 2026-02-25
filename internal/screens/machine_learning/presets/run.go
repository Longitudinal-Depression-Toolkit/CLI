package presets

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"ldt-toolkit-cli/internal/screens/machine_learning/support/schema"
	"ldt-toolkit-cli/internal/screens/machine_learning/support/ui"
	"ldt-toolkit-cli/internal/shared/components"
	"ldt-toolkit-cli/internal/shared/theme"
)

const catalogOperation = "machine_learning.presets.catalog"

func Catalog(execute BridgeExecutor) ([]Preset, error) {
	if execute == nil {
		return nil, errors.New("machine learning presets runtime has no bridge executor")
	}
	result, err := execute(catalogOperation, map[string]any{})
	if err != nil {
		return nil, err
	}
	return schema.DecodePresetPayload(result)
}

func Run(preset Preset, runtime Runtime) error {
	module, ok := ResolveModule(preset)
	if !ok {
		return renderPresetNotice(Default(preset), runtime)
	}

	spec := module.Build(preset)
	if module.Run != nil && IsRunnable(preset) {
		err := module.Run(preset, runtime)
		if IsFlowCancelled(err) {
			return nil
		}
		return err
	}

	if !IsRunnable(preset) {
		title := strings.TrimSpace(spec.Title)
		if title == "" {
			title = strings.TrimSpace(preset.Name)
		}
		spec.Title = title
		spec.Summary = fmt.Sprintf("Preset status: %s", StatusLabel(preset))
		spec.Lines = []string{
			"This preset is incoming. Use tools while this workflow is being integrated.",
		}
	}
	return renderPresetNotice(spec, runtime)
}

func renderPresetNotice(spec ModuleSpec, runtime Runtime) error {
	ui.PrepareActionScreen("Machine Learning Presets", spec.Summary, runtime.InNavigator)
	components.PrintBlankLine()
	components.PrintLine(theme.App.SubtitleStyle().Render(strings.TrimSpace(spec.Title)))
	for _, line := range spec.Lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		components.PrintLine(theme.App.MutedTextStyle().Render(line))
	}
	return ui.RunExitCountdown("Returning to toolkit", 10*time.Second, runtime.InNavigator)
}
