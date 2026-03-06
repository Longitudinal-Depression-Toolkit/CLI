package datapp

import (
	"fmt"
	"strings"

	presetsmodule "ldt-toolkit-cli/internal/screens/data_preprocessing/presets"
	"ldt-toolkit-cli/internal/shared/components"
	"ldt-toolkit-cli/internal/shared/model"
	"ldt-toolkit-cli/internal/shared/theme"
)

func HandleBridgePath(path []string) (bool, error) {
	pathWithoutHelp, helpRequested := StripHelpFlags(path)

	canonical, handled := CanonicalPath(pathWithoutHelp)
	if !handled {
		return false, nil
	}
	if helpRequested {
		return true, renderDataPreprocessingHelp(canonical)
	}

	if tool, ok := ToolFromPath(canonical); ok {
		listOnly := strings.TrimSpace(tool.ListCommand) != "" && ContainsToken(canonical, tool.ListCommand)
		return true, runDataPreprocessingToolFlow(tool, listOnly)
	}
	if len(canonical) >= 3 && canonical[1] == "presets" {
		return true, runDataPreprocessingPresetFlowByID(canonical[2])
	}

	switch {
	case len(canonical) == 1:
		return true, runDataPreprocessingHub()
	case len(canonical) == 2 && canonical[1] == "tools":
		return true, runDataPreprocessingToolsHub()
	case len(canonical) == 2 && canonical[1] == "presets":
		return true, runDataPreprocessingPresetsHub()
	default:
		return true, fmt.Errorf(
			"unsupported data preprocessing route in Go CLI: %s",
			strings.Join(path, " "),
		)
	}
}

func renderDataPreprocessingHelp(canonical []string) error {
	if tool, ok := ToolFromPath(canonical); ok {
		return runDataPreprocessingToolFlow(tool, true)
	}
	if len(canonical) >= 3 && canonical[1] == "presets" {
		return runDataPreprocessingPresetFlowByID(canonical[2])
	}

	cfg := CurrentConfig()

	if len(canonical) >= 2 && canonical[1] == "tools" {
		components.PrintBlankLine()
		components.PrintLine(theme.App.SubtitleStyle().Render("Data preprocessing tools"))
		components.PrintLine(theme.App.MutedTextStyle().Render(cfg.ToolsSummary))
		components.PrintBlankLine()
		for _, command := range ToolCommands() {
			components.PrintLine("- " + model.CommandLabel(command) + ": " + command.Description)
		}
		return nil
	}

	if len(canonical) >= 2 && canonical[1] == "presets" {
		presets, err := presetsForDisplay()
		if err != nil {
			return err
		}
		components.PrintBlankLine()
		components.PrintLine(theme.App.SubtitleStyle().Render("Data preprocessing presets"))
		components.PrintLine(theme.App.MutedTextStyle().Render(cfg.PresetsSummary))
		components.PrintBlankLine()
		for _, preset := range presets {
			line := fmt.Sprintf("- %s: %s [%s]", preset.Name, preset.Description, presetsmodule.StatusLabel(preset))
			components.PrintLine(line)
		}
		return nil
	}

	components.PrintBlankLine()
	components.PrintLine(theme.App.SubtitleStyle().Render("Data preprocessing"))
	components.PrintLine(theme.App.MutedTextStyle().Render(cfg.RootSummary))
	components.PrintBlankLine()
	for _, command := range ToolCommands() {
		components.PrintLine("- tools " + model.CommandLabel(command))
	}
	components.PrintLine("- presets")
	return nil
}

func LoadNode(path []string) (*parsedHelp, bool, error) {
	canonical, handled := CanonicalPath(path)
	if !handled {
		return nil, false, nil
	}

	usage := strings.Join(append([]string{usageCommandPrefix}, canonical...), " ")

	newNode := func(summary string, commands []commandDef) *parsedHelp {
		return &parsedHelp{
			Path:     model.ClonePath(path),
			Usage:    usage,
			Summary:  summary,
			Commands: commands,
			Raw:      "",
		}
	}

	cfg := CurrentConfig()

	switch {
	case len(canonical) == 1:
		return newNode(
			cfg.RootSummary,
			[]commandDef{
				{Name: "tools", DisplayName: "Tools", Description: "Primary data preprocessing tools for new studies."},
				{Name: "presets", DisplayName: "Presets", Description: "Reproducibility presets for published studies."},
			},
		), true, nil

	case len(canonical) == 2 && canonical[1] == "tools":
		return newNode(
			cfg.ToolsSummary,
			ToolCommands(),
		), true, nil

	case len(canonical) == 2 && canonical[1] == "presets":
		presets, err := presetsForDisplay()
		if err != nil {
			return nil, true, err
		}
		commands := make([]commandDef, 0, len(presets))
		for _, preset := range presets {
			description := strings.TrimSpace(preset.Description)
			if description == "" {
				description = strings.TrimSpace(preset.Name)
			}
			if !presetsmodule.IsRunnable(preset) {
				description = strings.TrimSpace(description + " [" + presetsmodule.StatusLabel(preset) + "]")
			}
			commands = append(commands, commandDef{
				Name:        preset.ID,
				DisplayName: strings.TrimSpace(preset.Name),
				Description: description,
			})
		}
		return newNode(
			cfg.PresetsSummary,
			commands,
		), true, nil

	case len(canonical) == 3 && canonical[1] == "presets":
		preset, ok, err := presetByIDForDisplay(canonical[2])
		if err != nil {
			return nil, true, err
		}
		if ok {
			return newNode(
				strings.TrimSpace(preset.Description),
				nil,
			), true, nil
		}
		return newNode(
			"Run configured preset in Go CLI.",
			nil,
		), true, nil
	}

	if tool, ok := ToolFromPath(canonical); ok {
		summary := strings.TrimSpace(tool.NodeSummary)
		if summary == "" {
			summary = strings.TrimSpace(tool.Description)
		}
		return newNode(summary, nil), true, nil
	}

	return nil, true, fmt.Errorf("unknown data preprocessing path: %s", strings.Join(path, " "))
}
