package preparemcsbyleap

type BridgeExecutor func(operation string, params map[string]any) (map[string]any, error)

type Runtime struct {
	Execute     BridgeExecutor
	InNavigator func() bool
}

type presetProfile struct {
	AvailableWaves []string            `json:"available_waves"`
	Defaults       presetProfileConfig `json:"defaults"`
}

type presetProfileConfig struct {
	ShowSummaryLogs         bool   `json:"show_summary_logs"`
	RunParallelWhenPossible bool   `json:"run_parallel_when_possible"`
	DefaultLongOutputPath   string `json:"default_long_output_path"`
	DefaultWideOutputPath   string `json:"default_wide_output_path"`
	DefaultWaveOutputDir    string `json:"default_wave_output_dir"`
	DefaultWideSuffixPrefix string `json:"default_wide_suffix_prefix"`
}

type presetRunResult struct {
	Preset          string                 `json:"preset"`
	Waves           []string               `json:"waves"`
	OutputFormat    string                 `json:"output_format"`
	SaveWaveOutputs bool                   `json:"save_wave_outputs"`
	SaveFinalOutput bool                   `json:"save_final_output"`
	Parallel        bool                   `json:"parallel"`
	MaxWorkers      *int                   `json:"max_workers"`
	LongOutputPath  string                 `json:"long_output_path"`
	WideOutputPath  string                 `json:"wide_output_path"`
	WaveOutputPaths []presetWaveOutputPath `json:"wave_output_paths"`
	LongShape       presetShape            `json:"long_shape"`
	WideShape       presetShape            `json:"wide_shape"`
}

type presetWaveOutputPath struct {
	Wave string `json:"wave"`
	Path string `json:"path"`
}

type presetShape struct {
	Rows    int `json:"rows"`
	Columns int `json:"columns"`
}
