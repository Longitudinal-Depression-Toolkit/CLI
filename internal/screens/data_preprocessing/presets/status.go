package presets

import "strings"

type Preset struct {
	ID          string
	Name        string
	Description string
	Status      string
}

func StatusLabel(preset Preset) string {
	status := strings.TrimSpace(preset.Status)
	if status == "" {
		return "available"
	}
	return status
}

func IsRunnable(preset Preset) bool {
	switch strings.ToLower(StatusLabel(preset)) {
	case "coming soon", "incoming", "planned", "todo", "wip", "disabled", "unavailable", "archived":
		return false
	default:
		return true
	}
}
