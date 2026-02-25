package dataprep

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"

	presetsmodule "ldt-toolkit-cli/internal/screens/data_preparation/presets"
	toolsmodule "ldt-toolkit-cli/internal/screens/data_preparation/tools"
	"ldt-toolkit-cli/internal/shared/components"
	"ldt-toolkit-cli/internal/shared/theme"
)

func runDataPreparationHub() error {
	var choice string
	form := components.NewLDTForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Data preparation view").
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
		return runDataPreparationToolsHub()
	case "presets":
		return runDataPreparationPresetsHub()
	default:
		return errors.New("no data preparation section selected")
	}
}

func runDataPreparationToolsHub() error {
	tools := Tools()
	if len(tools) == 0 {
		return errors.New("no data preparation tools were found")
	}

	options := make([]huh.Option[string], 0, len(tools))
	for _, tool := range tools {
		label := strings.TrimSpace(tool.Name)
		if label == "" {
			label = "Unnamed tool"
		}
		options = append(options, huh.NewOption(label, tool.ID))
	}

	choice := tools[0].ID
	form := components.NewLDTForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Data preparation tools").
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
		return errors.New("no data preparation tool selected")
	}
	return runDataPreparationToolFlow(tool, false)
}

func runDataPreparationPresetsHub() error {
	presets, err := presetsForDisplay()
	if err != nil {
		return err
	}
	if len(presets) == 0 {
		return errors.New("no data preparation presets were found")
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
				Title("Data preparation presets").
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

	return runDataPreparationPresetFlowByID(presetID)
}

func runDataPreparationPresetFlowByID(presetID string) error {
	selected, found, err := presetByIDForDisplay(presetID)
	if err != nil {
		return err
	}
	if !found {
		selected = presetsmodule.Preset{ID: presetID, Name: "Unnamed preset"}
		if configured, ok := PresetByID(presetID); ok {
			selected.Name = configured.Name
			selected.Description = configured.Description
		}
	}

	return presetsmodule.Run(
		selected,
		presetsmodule.Runtime{
			Execute:     executeBridge,
			InNavigator: currentNavigatorState,
		},
	)
}
