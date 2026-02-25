package presets

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"ldt-toolkit-cli/internal/screens/data_preparation/support/schema"
	"ldt-toolkit-cli/internal/screens/data_preparation/support/ui"
	"ldt-toolkit-cli/internal/shared/components"
	"ldt-toolkit-cli/internal/shared/theme"
)

const catalogOperation = "data_preparation.presets.catalog"

func Catalog(execute BridgeExecutor) ([]Preset, error) {
	if execute == nil {
		return nil, errors.New("data preparation presets runtime has no bridge executor")
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
		status := StatusLabel(preset)
		spec.Summary = fmt.Sprintf("Preset status: %s", status)
		spec.Lines = []string{
			fmt.Sprintf("This preset is currently %s.", status),
			"Select another preset or use tools mode while this one is in progress.",
		}
	}
	return renderPresetNotice(spec, runtime)
}

func renderPresetNotice(spec ModuleSpec, runtime Runtime) error {
	ui.PrepareActionScreen("Data Preparation Presets", spec.Summary, runtime.InNavigator)
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
