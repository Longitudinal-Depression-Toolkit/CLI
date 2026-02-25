package preparemcsbyleap

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"

	"ldt-toolkit-cli/internal/shared/components"
)

func PromptRunParams(profile presetProfile) (map[string]any, error) {
	wavesRaw := "ALL"
	wavesPrompt := components.NewLDTForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Waves to include").
				Description(fmt.Sprintf("Comma-separated (%s) or ALL.", strings.Join(profile.AvailableWaves, ", "))).
				Value(&wavesRaw),
		),
	)
	if err := wavesPrompt.Run(); err != nil {
		return nil, err
	}

	selectedWaves, err := parseWaveSelection(wavesRaw, profile.AvailableWaves)
	if err != nil {
		return nil, err
	}

	waveInputs := make(map[string]any, len(selectedWaves))
	for _, wave := range selectedWaves {
		confirmedPath, err := promptWaveInputPath(wave)
		if err != nil {
			return nil, err
		}
		waveInputs[wave] = confirmedPath
	}

	outputFormat := "long_and_wide"
	showSummaryLogs := profile.Defaults.ShowSummaryLogs
	saveFinalOutput := true
	saveWaveOutputs := len(selectedWaves) > 1
	parallel := len(selectedWaves) > 1 && profile.Defaults.RunParallelWhenPossible
	maxWorkersRaw := "auto"
	wideSuffixPrefix := strings.TrimSpace(profile.Defaults.DefaultWideSuffixPrefix)
	if wideSuffixPrefix == "" {
		wideSuffixPrefix = "_w"
	}

	generalForm := components.NewLDTForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Final output format").
				Options(
					huh.NewOption("Long", "long"),
					huh.NewOption("Wide", "wide"),
					huh.NewOption("Long and Wide", "long_and_wide"),
				).
				Value(&outputFormat),

			huh.NewConfirm().
				Title("Show stage summary logs?").
				Value(&showSummaryLogs),

			huh.NewConfirm().
				Title("Save final output datasets?").
				Value(&saveFinalOutput),
		),
	)
	if err := generalForm.Run(); err != nil {
		return nil, err
	}

	if len(selectedWaves) > 1 {
		parallelAndWaveForm := components.NewLDTForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Save wave-by-wave prepared outputs?").
					Value(&saveWaveOutputs),
				huh.NewConfirm().
					Title("Run wave preparation in parallel?").
					Value(&parallel),
			),
		)
		if err := parallelAndWaveForm.Run(); err != nil {
			return nil, err
		}
	}

	if parallel {
		maxWorkersMode := "auto"
		maxWorkersValue := 1
		parsedWorkers, parseErr := parseMaxWorkers(maxWorkersRaw)
		if parseErr == nil && parsedWorkers != nil {
			if workers, ok := parsedWorkers.(int); ok && workers > 0 {
				maxWorkersMode = "custom"
				maxWorkersValue = workers
			}
		}

		maxWorkersModeForm := components.NewLDTForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Max parallel workers").
					Description("Choose auto or set a custom worker count.").
					Options(
						huh.NewOption("Auto", "auto"),
						huh.NewOption("Custom", "custom"),
					).
					Value(&maxWorkersMode),
			),
		)
		if err := maxWorkersModeForm.Run(); err != nil {
			return nil, err
		}

		if maxWorkersMode == "custom" {
			minWorkers := 1.0
			workers, err := components.PromptIntStepper(components.NumberStepperConfig{
				Title:       "Max parallel workers",
				Description: "Use up/down arrows to set worker count.",
				Initial:     float64(maxWorkersValue),
				Step:        1,
				Min:         &minWorkers,
				Precision:   0,
				InputWidth:  18,
			})
			if err != nil {
				return nil, err
			}
			maxWorkersRaw = strconv.Itoa(workers)
		} else {
			maxWorkersRaw = "auto"
		}
	}

	if outputFormat == "wide" || outputFormat == "long_and_wide" {
		suffixForm := components.NewLDTForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Wide suffix prefix").
					Description("Final wide columns append this prefix + wave number.").
					Value(&wideSuffixPrefix).
					Validate(func(value string) error {
						if strings.TrimSpace(value) == "" {
							return errors.New("suffix prefix cannot be empty")
						}
						return nil
					}),
			),
		)
		if err := suffixForm.Run(); err != nil {
			return nil, err
		}
	}

	waveOutputDir := strings.TrimSpace(profile.Defaults.DefaultWaveOutputDir)
	if saveWaveOutputs {
		waveOutputForm := components.NewLDTForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Wave output directory").
					Description("Output directory for per-wave prepared CSVs.").
					Value(&waveOutputDir).
					Validate(func(value string) error {
						if strings.TrimSpace(value) == "" {
							return errors.New("output directory is required")
						}
						return nil
					}),
			),
		)
		if err := waveOutputForm.Run(); err != nil {
			return nil, err
		}
	}

	longOutputPath := strings.TrimSpace(profile.Defaults.DefaultLongOutputPath)
	if saveFinalOutput && (outputFormat == "long" || outputFormat == "long_and_wide") {
		longForm := components.NewLDTForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Final long output CSV path").
					Value(&longOutputPath).
					Validate(func(value string) error {
						if strings.TrimSpace(value) == "" {
							return errors.New("long output path is required")
						}
						return nil
					}),
			),
		)
		if err := longForm.Run(); err != nil {
			return nil, err
		}
	}

	wideOutputPath := strings.TrimSpace(profile.Defaults.DefaultWideOutputPath)
	if saveFinalOutput && (outputFormat == "wide" || outputFormat == "long_and_wide") {
		wideForm := components.NewLDTForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Final wide output CSV path").
					Value(&wideOutputPath).
					Validate(func(value string) error {
						if strings.TrimSpace(value) == "" {
							return errors.New("wide output path is required")
						}
						return nil
					}),
			),
		)
		if err := wideForm.Run(); err != nil {
			return nil, err
		}
	}

	maxWorkers, err := parseMaxWorkers(maxWorkersRaw)
	if err != nil {
		return nil, err
	}

	runParams := map[string]any{
		"waves":              strings.Join(selectedWaves, ","),
		"wave_inputs":        waveInputs,
		"output_format":      outputFormat,
		"wide_suffix_prefix": strings.TrimSpace(wideSuffixPrefix),
		"show_summary_logs":  showSummaryLogs,
		"save_wave_outputs":  saveWaveOutputs,
		"save_final_output":  saveFinalOutput,
		"parallel":           parallel,
		"max_workers":        maxWorkers,
	}
	if saveWaveOutputs {
		runParams["wave_output_dir"] = strings.TrimSpace(waveOutputDir)
	}
	if saveFinalOutput && (outputFormat == "long" || outputFormat == "long_and_wide") {
		runParams["long_output_path"] = strings.TrimSpace(longOutputPath)
	}
	if saveFinalOutput && (outputFormat == "wide" || outputFormat == "long_and_wide") {
		runParams["wide_output_path"] = strings.TrimSpace(wideOutputPath)
	}

	return runParams, nil
}

func promptWaveInputPath(wave string) (string, error) {
	selectedPath, err := components.PickPathWithFilePicker(
		fmt.Sprintf("%s raw Stata folder", wave),
		"Select the raw folder for this wave.",
		components.PathPickerDirectory,
		"",
	)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(selectedPath), nil
}

func parseWaveSelection(raw string, available []string) ([]string, error) {
	if len(available) == 0 {
		return nil, errors.New("no waves available")
	}

	availableSet := make(map[string]struct{}, len(available))
	orderedAvailable := make([]string, 0, len(available))
	for _, wave := range available {
		token := strings.ToUpper(strings.TrimSpace(wave))
		if token == "" {
			continue
		}
		if _, seen := availableSet[token]; seen {
			continue
		}
		availableSet[token] = struct{}{}
		orderedAvailable = append(orderedAvailable, token)
	}

	normalized := strings.ToUpper(strings.TrimSpace(raw))
	if normalized == "" || normalized == "ALL" {
		return orderedAvailable, nil
	}

	tokens := strings.Split(raw, ",")
	selected := make([]string, 0, len(tokens))
	seen := make(map[string]struct{}, len(tokens))
	for _, token := range tokens {
		wave := strings.ToUpper(strings.TrimSpace(token))
		if wave == "" {
			continue
		}
		if _, ok := availableSet[wave]; !ok {
			return nil, fmt.Errorf("unsupported wave `%s`", wave)
		}
		if _, ok := seen[wave]; ok {
			continue
		}
		seen[wave] = struct{}{}
		selected = append(selected, wave)
	}
	if len(selected) == 0 {
		return nil, errors.New("at least one wave is required")
	}
	return selected, nil
}

func parseMaxWorkers(raw string) (any, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" || normalized == "auto" || normalized == "none" {
		return nil, nil
	}
	parsed, err := strconv.Atoi(normalized)
	if err != nil || parsed <= 0 {
		return nil, errors.New("max workers must be `auto` or a positive integer")
	}
	return parsed, nil
}
