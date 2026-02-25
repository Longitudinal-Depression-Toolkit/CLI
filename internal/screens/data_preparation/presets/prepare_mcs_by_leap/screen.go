package preparemcsbyleap

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"ldt-toolkit-cli/internal/screens/data_preparation/presets/spec"
	"ldt-toolkit-cli/internal/screens/data_preparation/support/ui"
	"ldt-toolkit-cli/internal/shared/components"
	"ldt-toolkit-cli/internal/shared/theme"
)

const (
	profileOperation = "data_preparation.presets.prepare_mcs_by_leap.profile"
	runOperation     = "data_preparation.presets.prepare_mcs_by_leap.run"
)

func Build(preset spec.Preset) spec.ModuleSpec {
	base := spec.Default(preset)
	base.Summary = "Run the five-stage MCS preset with guided path and output prompts."
	base.Lines = []string{
		"Collect raw wave directories and output options from the Go CLI.",
		"Execute the preset through Python library operations via bridge runner.",
	}
	return base
}

func Run(preset spec.Preset, runtime Runtime) error {
	if runtime.Execute == nil {
		return errors.New("prepare_MCS_by_LEAP runtime has no bridge executor")
	}

	profile, err := loadProfile(runtime.Execute)
	if err != nil {
		return err
	}
	if len(profile.AvailableWaves) == 0 {
		return errors.New("prepare_MCS_by_LEAP has no available waves configured")
	}

	subtitle := strings.TrimSpace(preset.Description)
	if subtitle == "" {
		subtitle = "Configure and run the five-stage MCS preset pipeline."
	}
	ui.PrepareActionScreen("Prepare MCS by LEAP", subtitle, runtime.InNavigator)

	runParams, err := PromptRunParams(profile)
	if err != nil {
		return err
	}

	rawResult, err := runtime.Execute(runOperation, map[string]any{"params": runParams})
	if err != nil {
		return err
	}

	result, err := decodeRunResult(rawResult)
	if err != nil {
		return err
	}

	printRunSummary(result)
	return ui.RunExitCountdown("Returning to toolkit", 10*time.Second, runtime.InNavigator)
}

func loadProfile(execute BridgeExecutor) (presetProfile, error) {
	raw, err := execute(profileOperation, map[string]any{})
	if err != nil {
		return presetProfile{}, err
	}

	payload, err := decodeProfile(raw)
	if err != nil {
		return presetProfile{}, err
	}
	return payload, nil
}

func decodeProfile(raw map[string]any) (presetProfile, error) {
	encoded, err := json.Marshal(raw)
	if err != nil {
		return presetProfile{}, fmt.Errorf("failed to encode preset profile payload: %w", err)
	}
	var payload presetProfile
	if err := json.Unmarshal(encoded, &payload); err != nil {
		return presetProfile{}, fmt.Errorf("failed to decode preset profile payload: %w", err)
	}
	return payload, nil
}

func decodeRunResult(raw map[string]any) (presetRunResult, error) {
	encoded, err := json.Marshal(raw)
	if err != nil {
		return presetRunResult{}, fmt.Errorf("failed to encode preset run payload: %w", err)
	}
	var payload presetRunResult
	if err := json.Unmarshal(encoded, &payload); err != nil {
		return presetRunResult{}, fmt.Errorf("failed to decode preset run payload: %w", err)
	}
	return payload, nil
}

func printRunSummary(result presetRunResult) {
	components.PrintBlankLine()
	components.PrintLine(theme.App.SubtitleStyle().Render("Prepare MCS by LEAP complete"))
	if len(result.Waves) > 0 {
		components.PrintfLine("- Waves: %s", strings.Join(result.Waves, ", "))
	}
	if strings.TrimSpace(result.OutputFormat) != "" {
		components.PrintfLine("- Output format: %s", outputFormatLabel(result.OutputFormat))
	}
	components.PrintfLine("- Parallel: %t", result.Parallel)
	if result.MaxWorkers != nil {
		components.PrintfLine("- Max workers: %d", *result.MaxWorkers)
	} else {
		components.PrintLine("- Max workers: auto")
	}

	if strings.TrimSpace(result.LongOutputPath) != "" {
		components.PrintfLine("- Long output: %s", result.LongOutputPath)
	}
	if strings.TrimSpace(result.WideOutputPath) != "" {
		components.PrintfLine("- Wide output: %s", result.WideOutputPath)
	}
	for _, item := range result.WaveOutputPaths {
		components.PrintfLine("- Wave %s output: %s", item.Wave, item.Path)
	}
	if result.LongShape.Rows > 0 || result.LongShape.Columns > 0 {
		components.PrintfLine("- Long shape: %d rows x %d columns", result.LongShape.Rows, result.LongShape.Columns)
	}
	if result.WideShape.Rows > 0 || result.WideShape.Columns > 0 {
		components.PrintfLine("- Wide shape: %d rows x %d columns", result.WideShape.Rows, result.WideShape.Columns)
	}
}

func outputFormatLabel(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "long":
		return "Long"
	case "wide":
		return "Wide"
	case "both":
		return "Long and Wide"
	default:
		return strings.TrimSpace(raw)
	}
}
