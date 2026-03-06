package datapp

import (
	"sort"
	"strings"

	presetsmodule "ldt-toolkit-cli/internal/screens/data_preprocessing/presets"
)

func presetsForDisplay() ([]presetsmodule.Preset, error) {
	presets, err := presetsmodule.Catalog(executeBridge)
	if err != nil {
		fallback := configuredPresetsAsRuntime()
		if len(fallback) > 0 {
			return fallback, nil
		}
		return nil, err
	}
	return mergePresetsWithConfigured(presets), nil
}

func presetByIDForDisplay(presetID string) (presetsmodule.Preset, bool, error) {
	presets, err := presetsForDisplay()
	if err != nil {
		return presetsmodule.Preset{}, false, err
	}
	for _, preset := range presets {
		if canonicalToken(preset.ID) == canonicalToken(presetID) {
			return preset, true, nil
		}
	}
	return presetsmodule.Preset{}, false, nil
}

func configuredPresetsAsRuntime() []presetsmodule.Preset {
	configured := Presets()
	result := make([]presetsmodule.Preset, 0, len(configured))
	for _, preset := range configured {
		name := strings.TrimSpace(preset.Name)
		if name == "" {
			name = "Unnamed preset"
		}
		description := strings.TrimSpace(preset.Description)
		if description == "" {
			description = name
		}
		status := strings.TrimSpace(preset.Status)
		if status == "" {
			status = "available"
		}
		available := true
		if strings.EqualFold(status, "coming_soon") || strings.EqualFold(status, "coming soon") ||
			strings.EqualFold(status, "incoming") || strings.EqualFold(status, "planned") ||
			strings.EqualFold(status, "todo") || strings.EqualFold(status, "wip") ||
			strings.EqualFold(status, "disabled") || strings.EqualFold(status, "unavailable") {
			available = false
		}

		result = append(result, presetsmodule.Preset{
			ID:          preset.ID,
			Name:        name,
			Description: description,
			Available:   available,
			Status:      status,
		})
	}
	sort.SliceStable(result, func(i, j int) bool {
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})
	return result
}

func mergePresetsWithConfigured(runtimePresets []presetsmodule.Preset) []presetsmodule.Preset {
	configured := Presets()
	byID := make(map[string]PresetConfig, len(configured))
	for _, preset := range configured {
		byID[canonicalToken(preset.ID)] = preset
	}

	merged := make([]presetsmodule.Preset, 0, len(runtimePresets))
	for _, preset := range runtimePresets {
		canonicalID := canonicalToken(preset.ID)
		cfg, hasCfg := byID[canonicalID]

		name := strings.TrimSpace(preset.Name)
		if name == "" && hasCfg {
			name = strings.TrimSpace(cfg.Name)
		}
		if name == "" {
			name = "Unnamed preset"
		}

		description := strings.TrimSpace(preset.Description)
		if description == "" && hasCfg {
			description = strings.TrimSpace(cfg.Description)
		}
		if description == "" {
			description = name
		}

		status := strings.TrimSpace(preset.Status)
		if status == "" && hasCfg {
			status = strings.TrimSpace(cfg.Status)
		}
		if status == "" && !preset.Available {
			status = "coming soon"
		}
		if status == "" && preset.Available {
			status = "available"
		}

		merged = append(merged, presetsmodule.Preset{
			ID:          preset.ID,
			Name:        name,
			Description: description,
			Available:   preset.Available,
			Status:      status,
		})
	}

	sort.SliceStable(merged, func(i, j int) bool {
		return strings.ToLower(merged[i].Name) < strings.ToLower(merged[j].Name)
	})
	return merged
}
