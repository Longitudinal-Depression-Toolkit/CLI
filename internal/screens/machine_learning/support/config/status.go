package config

import "strings"

func StatusLabel(raw string) string {
	status := strings.TrimSpace(raw)
	if status == "" {
		return "available"
	}
	return status
}

func IsRunnableStatus(raw string) bool {
	switch strings.ToLower(StatusLabel(raw)) {
	case "coming soon", "incoming", "planned", "todo", "wip", "disabled", "unavailable", "archived":
		return false
	default:
		return true
	}
}
