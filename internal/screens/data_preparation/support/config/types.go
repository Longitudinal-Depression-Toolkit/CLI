package config

const (
	dataPreparationConfigEnvVar      = "LDT_DATA_PREPARATION_CONFIG"
	defaultDataPreparationConfigPath = "config/data_preparation.json"
)

type ToolConfig struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	NodeSummary      string `json:"node_summary"`
	Subtitle         string `json:"subtitle"`
	TableTitle       string `json:"table_title"`
	SelectionTitle   string `json:"selection_title"`
	CatalogOperation string `json:"catalog_operation"`
	RunOperation     string `json:"run_operation"`
	ListCommand      string `json:"list_command"`
}

type PresetConfig struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type Config struct {
	RootSummary    string         `json:"root_summary"`
	ToolsSummary   string         `json:"tools_summary"`
	PresetsSummary string         `json:"presets_summary"`
	Tools          []ToolConfig   `json:"tools"`
	Presets        []PresetConfig `json:"presets"`
}

var currentConfig Config

func cloneConfig(cfg Config) Config {
	copied := cfg
	copied.Tools = append([]ToolConfig(nil), cfg.Tools...)
	copied.Presets = append([]PresetConfig(nil), cfg.Presets...)
	return copied
}
