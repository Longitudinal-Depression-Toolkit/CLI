package model

import "strings"

type CommandDef struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

type ParsedHelp struct {
	Path     []string
	Usage    string
	Summary  string
	Commands []CommandDef
	Raw      string
}

func ClonePath(path []string) []string {
	copied := make([]string, len(path))
	copy(copied, path)
	return copied
}

func JoinPath(path []string) string {
	return strings.Join(path, "\x1f")
}

func IntMax(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func IntMin(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func ParseCSVValues(raw string) []string {
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		candidate := strings.TrimSpace(part)
		if candidate == "" {
			continue
		}
		values = append(values, candidate)
	}
	return values
}

func CommandLabel(command CommandDef) string {
	display := strings.TrimSpace(command.DisplayName)
	if display != "" {
		return display
	}
	name := strings.TrimSpace(command.Name)
	if name != "" {
		return name
	}
	return "Unnamed"
}
