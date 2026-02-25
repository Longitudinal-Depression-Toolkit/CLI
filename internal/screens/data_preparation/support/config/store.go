package config

import (
	"strings"

	"ldt-toolkit-cli/internal/shared/model"
)

func Configure() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	currentConfig = cfg
	return nil
}

func CurrentConfig() Config {
	return cloneConfig(currentConfig)
}

func ToolCommands() []model.CommandDef {
	cfg := CurrentConfig()
	commands := make([]model.CommandDef, 0, len(cfg.Tools))
	for _, tool := range cfg.Tools {
		commands = append(commands, model.CommandDef{
			Name:        tool.ID,
			DisplayName: tool.Name,
			Description: tool.Description,
		})
	}
	return commands
}

func Tools() []ToolConfig {
	cfg := CurrentConfig()
	return cfg.Tools
}

func Presets() []PresetConfig {
	cfg := CurrentConfig()
	return cfg.Presets
}

func ToolByID(id string) (ToolConfig, bool) {
	needle := strings.TrimSpace(id)
	for _, tool := range currentConfig.Tools {
		if strings.EqualFold(tool.ID, needle) {
			return tool, true
		}
	}
	return ToolConfig{}, false
}

func ToolFromPath(path []string) (ToolConfig, bool) {
	for _, tool := range currentConfig.Tools {
		if containsToken(path, tool.ID) {
			return tool, true
		}
	}
	return ToolConfig{}, false
}

func PresetByID(id string) (PresetConfig, bool) {
	needle := strings.TrimSpace(id)
	for _, preset := range currentConfig.Presets {
		if strings.EqualFold(preset.ID, needle) {
			return preset, true
		}
	}
	return PresetConfig{}, false
}

func containsToken(path []string, token string) bool {
	for _, part := range path {
		if part == token {
			return true
		}
	}
	return false
}
