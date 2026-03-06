package presets

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/huh"

	"ldt-toolkit-cli/internal/screens/data_preprocessing/support/ui"
	"ldt-toolkit-cli/internal/shared/components"
	"ldt-toolkit-cli/internal/shared/theme"
)

const (
	preprocessMCSByLEAPID      = "preprocess_mcs_by_leap"
	profileOperationPreprocess = "data_preprocessing.presets.preprocess_mcs_by_leap.profile"
	runOperationPreprocess     = "data_preprocessing.presets.preprocess_mcs_by_leap.run"
)

type presetProfile struct {
	Preset   string              `json:"preset"`
	Defaults presetProfileConfig `json:"defaults"`
}

type presetProfileConfig struct {
	ShowSummaryLogs         bool   `json:"show_summary_logs"`
	SaveAuditTables         bool   `json:"save_audit_tables"`
	DefaultOutputPath       string `json:"default_output_path"`
	DefaultAuditOutputDir   string `json:"default_audit_output_dir"`
	DefaultStage0ConfigPath string `json:"default_stage_0_config_path"`
	DefaultStage1ConfigPath string `json:"default_stage_1_config_path"`
	DefaultStage2ConfigPath string `json:"default_stage_2_config_path"`
	DefaultStage3ConfigPath string `json:"default_stage_3_config_path"`
	DefaultStage4ConfigPath string `json:"default_stage_4_config_path"`
	DefaultStage5ConfigPath string `json:"default_stage_5_config_path"`
}

type presetRunResult struct {
	Preset          string                 `json:"preset"`
	InputPath       string                 `json:"input_path"`
	OutputPath      string                 `json:"output_path"`
	AuditOutputDir  string                 `json:"audit_output_dir"`
	SaveFinalOutput bool                   `json:"save_final_output"`
	SaveAuditTables bool                   `json:"save_audit_tables"`
	Shape           shapeSummary           `json:"shape"`
	Stage0Summary   map[string]any         `json:"stage_0_summary"`
	Stage1Summary   map[string]any         `json:"stage_1_summary"`
	Stage2Summary   map[string]any         `json:"stage_2_summary"`
	Stage3Summary   map[string]any         `json:"stage_3_summary"`
	Stage4Summary   map[string]any         `json:"stage_4_summary"`
	Stage5Summary   map[string]any         `json:"stage_5_summary"`
	AuditFiles      []presetRunResultAudit `json:"audit_files"`
}

type shapeSummary struct {
	Rows    int `json:"rows"`
	Columns int `json:"columns"`
}

type presetRunResultAudit struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func Run(preset Preset, runtime Runtime) error {
	if !IsRunnable(preset) {
		return renderIncomingPresetNotice(preset, runtime)
	}

	if !strings.EqualFold(strings.TrimSpace(preset.ID), preprocessMCSByLEAPID) {
		return renderUnwiredPresetNotice(preset, runtime)
	}

	if runtime.Execute == nil {
		return errors.New("preprocess_MCS_by_LEAP runtime has no bridge executor")
	}

	subtitle := strings.TrimSpace(preset.Description)
	if subtitle == "" {
		subtitle = "Configure and run the staged preprocessing preset."
	}
	ui.PrepareActionScreen("Preprocess MCS by LEAP", subtitle, runtime.InNavigator)
	printPreprocessMCSByLEAPGuideSummary()

	profile, err := loadProfile(runtime.Execute)
	if err != nil {
		return err
	}

	runParams, err := promptRunParams(profile)
	if err != nil {
		return err
	}

	rawResult, err := runtime.Execute(runOperationPreprocess, map[string]any{"params": runParams})
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

func renderIncomingPresetNotice(preset Preset, runtime Runtime) error {
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
	components.PrintLine(theme.App.MutedTextStyle().Render(fmt.Sprintf("Preset status: %s", status)))
	components.PrintLine(theme.App.MutedTextStyle().Render("This preset is incoming. Use tools while this workflow is being integrated."))

	return ui.RunExitCountdown("Returning to toolkit", 10*time.Second, runtime.InNavigator)
}

func renderUnwiredPresetNotice(preset Preset, runtime Runtime) error {
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
	components.PrintLine(theme.App.MutedTextStyle().Render("This preset is available but not wired yet in the Go runtime."))
	return ui.RunExitCountdown("Returning to toolkit", 10*time.Second, runtime.InNavigator)
}

func printPreprocessMCSByLEAPGuideSummary() {
	components.PrintBlankLine()
	components.PrintLine(theme.App.SubtitleStyle().Render("Preset guide"))
	components.PrintLine(theme.App.MutedTextStyle().Render("Requires the wide artefact from Prepare MCS by LEAP. Long-only output is not supported."))
	components.PrintLine(theme.App.MutedTextStyle().Render("Recommended input: data/processed/MCS/mcs_longitudinal_wide.csv"))
	components.PrintLine(theme.App.MutedTextStyle().Render("Final target lineage: EMOTION_w7"))
	components.PrintLine(theme.App.MutedTextStyle().Render("Pipeline flow: Stage 0 target rows -> Stage 1 structural -> Stage 2 sentinels -> Stage 3 leakage policy -> Stage 4 finalisation diagnostics -> Stage 5 encoding policy"))
	components.PrintLine(theme.App.MutedTextStyle().Render("Expected outputs: preprocessed wide CSV plus preprocess_logs audit tables."))
	components.PrintBlankLine()
}

func loadProfile(execute BridgeExecutor) (presetProfile, error) {
	raw, err := execute(profileOperationPreprocess, map[string]any{})
	if err != nil {
		return presetProfile{}, err
	}

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

func promptRunParams(profile presetProfile) (map[string]any, error) {
	inputPath, err := components.PickPathWithFilePicker(
		"Input prepared wide CSV",
		"Select the prepared MCS wide CSV to preprocess.",
		components.PathPickerFile,
		"",
	)
	if err != nil {
		return nil, err
	}

	showSummaryLogs := profile.Defaults.ShowSummaryLogs
	saveFinalOutput := true
	saveAuditTables := profile.Defaults.SaveAuditTables

	baseForm := components.NewLDTForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Show stage summary logs?").
				Value(&showSummaryLogs),
			huh.NewConfirm().
				Title("Save final preprocessed output CSV?").
				Value(&saveFinalOutput),
			huh.NewConfirm().
				Title("Save per-stage audit tables?").
				Value(&saveAuditTables),
		),
	)
	if err := baseForm.Run(); err != nil {
		return nil, err
	}

	outputPath := strings.TrimSpace(profile.Defaults.DefaultOutputPath)
	if saveFinalOutput {
		outputForm := components.NewLDTForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Final output CSV path").
					Value(&outputPath).
					Validate(func(value string) error {
						if strings.TrimSpace(value) == "" {
							return errors.New("output path is required")
						}
						return nil
					}),
			),
		)
		if err := outputForm.Run(); err != nil {
			return nil, err
		}
	}

	auditOutputDir := strings.TrimSpace(profile.Defaults.DefaultAuditOutputDir)
	if saveAuditTables {
		auditForm := components.NewLDTForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Audit output directory").
					Value(&auditOutputDir).
					Validate(func(value string) error {
						if strings.TrimSpace(value) == "" {
							return errors.New("audit output directory is required")
						}
						return nil
					}),
			),
		)
		if err := auditForm.Run(); err != nil {
			return nil, err
		}
	}

	customStageConfigs := false
	stage0 := strings.TrimSpace(profile.Defaults.DefaultStage0ConfigPath)
	stage1 := strings.TrimSpace(profile.Defaults.DefaultStage1ConfigPath)
	stage2 := strings.TrimSpace(profile.Defaults.DefaultStage2ConfigPath)
	stage3 := strings.TrimSpace(profile.Defaults.DefaultStage3ConfigPath)
	stage4 := strings.TrimSpace(profile.Defaults.DefaultStage4ConfigPath)
	stage5 := strings.TrimSpace(profile.Defaults.DefaultStage5ConfigPath)
	customConfigForm := components.NewLDTForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Override default stage config paths?").
				Value(&customStageConfigs),
		),
	)
	if err := customConfigForm.Run(); err != nil {
		return nil, err
	}
	if customStageConfigs {
		stageForm := components.NewLDTForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Stage 0 config path").
					Value(&stage0),
				huh.NewInput().
					Title("Stage 1 config path").
					Value(&stage1),
				huh.NewInput().
					Title("Stage 2 config path").
					Value(&stage2),
				huh.NewInput().
					Title("Stage 3 config path").
					Value(&stage3),
				huh.NewInput().
					Title("Stage 4 config path").
					Value(&stage4),
				huh.NewInput().
					Title("Stage 5 config path").
					Value(&stage5),
			),
		)
		if err := stageForm.Run(); err != nil {
			return nil, err
		}
	}

	params := map[string]any{
		"input_path":        strings.TrimSpace(inputPath),
		"show_summary_logs": showSummaryLogs,
		"save_final_output": saveFinalOutput,
		"save_audit_tables": saveAuditTables,
	}

	if saveFinalOutput {
		params["output_path"] = strings.TrimSpace(outputPath)
	}
	if saveAuditTables {
		params["audit_output_dir"] = strings.TrimSpace(auditOutputDir)
	}
	if customStageConfigs {
		params["stage_0_config_path"] = strings.TrimSpace(stage0)
		params["stage_1_config_path"] = strings.TrimSpace(stage1)
		params["stage_2_config_path"] = strings.TrimSpace(stage2)
		params["stage_3_config_path"] = strings.TrimSpace(stage3)
		params["stage_4_config_path"] = strings.TrimSpace(stage4)
		params["stage_5_config_path"] = strings.TrimSpace(stage5)
	}

	return params, nil
}

func printRunSummary(result presetRunResult) {
	components.PrintBlankLine()
	components.PrintLine(theme.App.SubtitleStyle().Render("Preprocess MCS by LEAP complete"))
	components.PrintfLine("- Input: %s", strings.TrimSpace(result.InputPath))
	if strings.TrimSpace(result.OutputPath) != "" {
		components.PrintfLine("- Output: %s", strings.TrimSpace(result.OutputPath))
	} else {
		components.PrintLine("- Output: (not saved)")
	}
	if strings.TrimSpace(result.AuditOutputDir) != "" {
		components.PrintfLine("- Audit directory: %s", strings.TrimSpace(result.AuditOutputDir))
	}
	components.PrintfLine("- Shape: %d rows x %d columns", result.Shape.Rows, result.Shape.Columns)
	components.PrintfLine("- Audit files: %d", len(result.AuditFiles))

	if v, ok := asInt(result.Stage0Summary["rows_dropped_missing_target"]); ok {
		components.PrintfLine("- Stage 0 target-missing rows dropped: %d", v)
	}
	if v, ok := asInt(result.Stage1Summary["dropped_columns"]); ok {
		components.PrintfLine("- Stage 1 dropped columns: %d", v)
	}
	if v, ok := asInt(result.Stage2Summary["total_cells_replaced"]); ok {
		components.PrintfLine("- Stage 2 sentinel cells replaced: %d", v)
	}
	if v, ok := asInt(result.Stage3Summary["leakage_columns_removed"]); ok {
		components.PrintfLine("- Stage 3 leakage columns removed: %d", v)
	}
	if v, ok := asInt(result.Stage4Summary["high_corr_columns_dropped"]); ok {
		components.PrintfLine("- Stage 4 high-correlation columns dropped: %d", v)
	}
	if v, ok := asInt(result.Stage5Summary["nominal_created_columns"]); ok {
		components.PrintfLine("- Stage 5 nominal columns created: %d", v)
	}
}

func asInt(value any) (int, bool) {
	switch typed := value.(type) {
	case float64:
		return int(typed), true
	case float32:
		return int(typed), true
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case int32:
		return int(typed), true
	default:
		return 0, false
	}
}
