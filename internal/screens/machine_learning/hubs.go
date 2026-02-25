package machinelearning

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"

	presetsmodule "ldt-toolkit-cli/internal/screens/machine_learning/presets"
	toolsmodule "ldt-toolkit-cli/internal/screens/machine_learning/tools"
	"ldt-toolkit-cli/internal/shared/components"
	"ldt-toolkit-cli/internal/shared/theme"
)

func runMachineLearningHub() error {
	var choice string
	form := components.NewLDTForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Machine learning view").
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
		return runMachineLearningToolsHub()
	case "presets":
		return runMachineLearningPresetsHub()
	default:
		return errors.New("no machine learning section selected")
	}
}

func runMachineLearningToolsHub() error {
	tools := Tools()
	if len(tools) == 0 {
		return errors.New("no machine learning tools were found")
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
				Title("Machine learning tools").
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
		return errors.New("no machine learning tool selected")
	}
	return runMachineLearningToolFlow(tool, false)
}

func runMachineLearningPresetsHub() error {
	presets, err := presetsForDisplay()
	if err != nil {
		return err
	}
	if len(presets) == 0 {
		return errors.New("no machine learning presets were found")
	}

	options := make([]huh.Option[string], 0, len(presets))
	for _, preset := range presets {
		label := strings.TrimSpace(preset.Name)
		if label == "" {
			label = "Unnamed preset"
		}
		if !presetsmodule.IsRunnable(preset) {
			label = theme.App.MutedTextStyle().Render(
				fmt.Sprintf("%s (%s)", label, presetsmodule.StatusLabel(preset)),
			)
		}
		options = append(options, huh.NewOption(label, preset.ID))
	}

	var presetID string
	form := components.NewLDTForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Machine learning presets").
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

	return runMachineLearningPresetFlowByID(presetID)
}

func runMachineLearningPresetFlowByID(presetID string) error {
	preset, ok, err := presetByIDForDisplay(presetID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("unknown machine learning preset: %s", strings.TrimSpace(presetID))
	}
	return presetsmodule.Run(
		preset,
		presetsmodule.Runtime{
			Execute:     executeBridge,
			InNavigator: currentNavigatorState,
		},
	)
}
