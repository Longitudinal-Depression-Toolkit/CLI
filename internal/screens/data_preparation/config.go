package dataprep

import (
	dpconfig "ldt-toolkit-cli/internal/screens/data_preparation/support/config"
	"ldt-toolkit-cli/internal/shared/model"
)

type ToolConfig = dpconfig.ToolConfig
type PresetConfig = dpconfig.PresetConfig
type Config = dpconfig.Config

func Configure() error {
	return dpconfig.Configure()
}

func CurrentConfig() Config {
	return dpconfig.CurrentConfig()
}

func ToolCommands() []model.CommandDef {
	return dpconfig.ToolCommands()
}

func Tools() []ToolConfig {
	return dpconfig.Tools()
}

func ToolByID(id string) (ToolConfig, bool) {
	return dpconfig.ToolByID(id)
}

func ToolFromPath(path []string) (ToolConfig, bool) {
	return dpconfig.ToolFromPath(path)
}

func Presets() []PresetConfig {
	return dpconfig.Presets()
}

func PresetByID(id string) (PresetConfig, bool) {
	return dpconfig.PresetByID(id)
}
