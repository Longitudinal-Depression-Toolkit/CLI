package presets

import "strings"

type Preset struct {
	ID          string
	Name        string
	Description string
	Available   bool
	Status      string
}

func StatusLabel(preset Preset) string {
	status := strings.TrimSpace(preset.Status)
	if status == "" {
		if !preset.Available {
			return "coming soon"
		}
		return "available"
	}
	return status
}

func IsRunnable(preset Preset) bool {
	switch strings.ToLower(StatusLabel(preset)) {
	case "coming soon", "incoming", "planned", "todo", "wip", "disabled", "unavailable", "archived":
		return false
	default:
		return preset.Available || strings.TrimSpace(preset.Status) != ""
	}
}
