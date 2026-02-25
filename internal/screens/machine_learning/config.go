package machinelearning

import (
	"strings"

	config "ldt-toolkit-cli/internal/screens/machine_learning/support/config"
	"ldt-toolkit-cli/internal/shared/model"
)

type commandDef = model.CommandDef
type parsedHelp = model.ParsedHelp

type ToolConfig = config.ToolConfig
type PresetConfig = config.PresetConfig

func Configure() error {
	return config.Configure()
}

func CurrentConfig() config.Config {
	return config.CurrentConfig()
}

func ToolCommands() []model.CommandDef {
	return config.ToolCommands()
}

func Tools() []ToolConfig {
	return config.Tools()
}

func ToolByID(id string) (ToolConfig, bool) {
	return config.ToolByID(id)
}

func ToolFromPath(path []string) (ToolConfig, bool) {
	return config.ToolFromPath(path)
}

func Presets() []PresetConfig {
	return config.Presets()
}

func PresetByID(id string) (PresetConfig, bool) {
	return config.PresetByID(id)
}

func ToolStatusLabel(tool ToolConfig) string {
	return config.StatusLabel(tool.Status)
}

func ToolIsRunnable(tool ToolConfig) bool {
	return config.IsRunnableStatus(tool.Status) &&
		strings.TrimSpace(tool.CatalogOperation) != "" &&
		strings.TrimSpace(tool.RunOperation) != ""
}

func PresetStatusLabel(preset PresetConfig) string {
	return config.StatusLabel(preset.Status)
}

func PresetIsRunnable(preset PresetConfig) bool {
	return config.IsRunnableStatus(preset.Status)
}
