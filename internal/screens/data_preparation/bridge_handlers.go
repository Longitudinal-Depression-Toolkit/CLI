package dataprep

import (
	"fmt"
	"strings"

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
		return true, renderDataPreparationHelp(canonical)
	}

	if tool, ok := ToolFromPath(canonical); ok {
		listOnly := strings.TrimSpace(tool.ListCommand) != "" && ContainsToken(canonical, tool.ListCommand)
		return true, runDataPreparationToolFlow(tool, listOnly)
	}
	if len(canonical) >= 3 && canonical[1] == "presets" {
		return true, runDataPreparationPresetFlowByID(canonical[2])
	}

	switch {
	case len(canonical) == 1:
		return true, runDataPreparationHub()
	case len(canonical) == 2 && canonical[1] == "tools":
		return true, runDataPreparationToolsHub()
	case len(canonical) == 2 && canonical[1] == "presets":
		return true, runDataPreparationPresetsHub()
	default:
		return true, fmt.Errorf(
			"unsupported data preparation route in Go CLI: %s",
			strings.Join(path, " "),
		)
	}
}

func renderDataPreparationHelp(canonical []string) error {
	if tool, ok := ToolFromPath(canonical); ok {
		return runDataPreparationToolFlow(tool, true)
	}
	if len(canonical) >= 3 && canonical[1] == "presets" {
		return runDataPreparationPresetFlowByID(canonical[2])
	}

	cfg := CurrentConfig()

	if len(canonical) >= 2 && canonical[1] == "tools" {
		components.PrintBlankLine()
		components.PrintLine(theme.App.SubtitleStyle().Render("Data preparation tools"))
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
		components.PrintLine(theme.App.SubtitleStyle().Render("Data preparation presets"))
		if len(presets) == 0 {
			components.PrintLine(theme.App.MutedTextStyle().Render("No presets configured."))
			return nil
		}
		for _, preset := range presets {
			label := strings.TrimSpace(preset.Name)
			if label == "" {
				label = "Unnamed preset"
			}
			line := fmt.Sprintf("- %s: %s", label, preset.Description)
			components.PrintLine(line)
		}
		return nil
	}

	components.PrintBlankLine()
	components.PrintLine(theme.App.SubtitleStyle().Render("Data preparation"))
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
				{Name: "tools", DisplayName: "Tools", Description: "Primary data preparation tools for new studies."},
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
			commands = append(commands, commandDef{
				Name:        preset.ID,
				DisplayName: strings.TrimSpace(preset.Name),
				Description: preset.Description,
			})
		}
		return newNode(
			cfg.PresetsSummary,
			commands,
		), true, nil

	case len(canonical) == 3 && canonical[1] == "presets":
		preset, found, err := presetByIDForDisplay(canonical[2])
		if err != nil {
			return nil, true, err
		}
		if found {
			return newNode(
				strings.TrimSpace(preset.Description),
				nil,
			), true, nil
		}
		if configured, ok := PresetByID(canonical[2]); ok {
			return newNode(
				strings.TrimSpace(configured.Description),
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

	return nil, true, fmt.Errorf("unknown data preparation path: %s", strings.Join(path, " "))
}
