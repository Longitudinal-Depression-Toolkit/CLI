package datapp

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"

	presetsmodule "ldt-toolkit-cli/internal/screens/data_preprocessing/presets"
	toolsmodule "ldt-toolkit-cli/internal/screens/data_preprocessing/tools"
	"ldt-toolkit-cli/internal/shared/components"
	"ldt-toolkit-cli/internal/shared/theme"
)

func runDataPreprocessingHub() error {
	var choice string
	form := components.NewLDTForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Data preprocessing view").
				Options(
					huh.NewOption("Tools", "tools"),
					huh.NewOption("Presets reproducibility", "presets"),
				).
				Value(&choice),
		),
	)

	if err := form.Run(); err != nil {
		if presetsmodule.IsFlowCancelled(err) {
			return nil
		}
		return err
	}

	switch choice {
	case "tools":
		return runDataPreprocessingToolsHub()
	case "presets":
		return runDataPreprocessingPresetsHub()
	default:
		return errors.New("no data preprocessing section selected")
	}
}

func runDataPreprocessingToolsHub() error {
	tools := Tools()
	if len(tools) == 0 {
		return errors.New("no data preprocessing tools were found")
	}

	options := make([]huh.Option[string], 0, len(tools))
	for _, tool := range tools {
		label := strings.TrimSpace(tool.Name)
		if label == "" {
			label = "Unnamed tool"
		}
		if !ToolIsRunnable(tool) {
			label = theme.App.MutedTextStyle().Render(
				fmt.Sprintf("%s (%s)", label, ToolStatusLabel(tool)),
			)
		}
		options = append(options, huh.NewOption(label, tool.ID))
	}

	choice := tools[0].ID
	form := components.NewLDTForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Data preprocessing tools").
				Options(options...).
				Value(&choice),
		),
	)

	if err := form.Run(); err != nil {
		if toolsmodule.IsFlowCancelled(err) {
			return nil
		}
		return err
	}

	tool, ok := ToolByID(choice)
	if !ok {
		return errors.New("no data preprocessing tool selected")
	}
	return runDataPreprocessingToolFlow(tool, false)
}

func runDataPreprocessingPresetsHub() error {
	presets := Presets()
	if len(presets) == 0 {
		return errors.New("no data preprocessing presets were found")
	}

	options := make([]huh.Option[string], 0, len(presets))
	for _, preset := range presets {
		label := strings.TrimSpace(preset.Name)
		if label == "" {
			label = "Unnamed preset"
		}
		if !PresetIsRunnable(preset) {
			label = theme.App.MutedTextStyle().Render(
				fmt.Sprintf("%s (%s)", label, PresetStatusLabel(preset)),
			)
		}
		options = append(options, huh.NewOption(label, preset.ID))
	}

	var presetID string
	form := components.NewLDTForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Data preprocessing presets").
				Options(options...).
				Value(&presetID),
		),
	)

	if err := form.Run(); err != nil {
		if presetsmodule.IsFlowCancelled(err) {
			return nil
		}
		return err
	}

	return runDataPreprocessingPresetFlowByID(presetID)
}

func runDataPreprocessingPresetFlowByID(presetID string) error {
	preset, ok := PresetByID(presetID)
	if !ok {
		return fmt.Errorf("unknown data preprocessing preset: %s", strings.TrimSpace(presetID))
	}
	return presetsmodule.Run(
		presetsmodule.Preset{
			ID:          preset.ID,
			Name:        preset.Name,
			Description: preset.Description,
			Status:      preset.Status,
		},
		presetsmodule.Runtime{
			InNavigator: currentNavigatorState,
		},
	)
}
