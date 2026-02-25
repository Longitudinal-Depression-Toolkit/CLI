package datapp

import (
	"strings"
)

func Heading(path []string) string {
	canonical, handled := CanonicalPath(path)
	if !handled || len(canonical) == 0 {
		return "Data Preprocessing"
	}

	switch {
	case len(canonical) == 1:
		return "Data Preprocessing"
	case len(canonical) == 2 && canonical[1] == "tools":
		return "Data Preprocessing / Tools"
	case len(canonical) == 2 && canonical[1] == "presets":
		return "Data Preprocessing / Presets"
	case len(canonical) >= 3 && canonical[1] == "presets":
		if preset, ok := PresetByID(canonical[2]); ok {
			name := strings.TrimSpace(preset.Name)
			if name != "" {
				return name
			}
		}
		return strings.TrimSpace(canonical[2])
	}

	if tool, ok := ToolFromPath(canonical); ok {
		return tool.Name
	}

	parts := make([]string, 0, len(canonical))
	for _, part := range canonical {
		parts = append(parts, strings.TrimSpace(part))
	}
	return strings.Join(parts, " / ")
}

func CanonicalPath(path []string) ([]string, bool) {
	if len(path) == 0 {
		return nil, false
	}

	root := canonicalToken(path[0])
	if root != "data_preprocessing" {
		return nil, false
	}

	canonical := make([]string, 0, len(path)+1)
	canonical = append(canonical, "data_preprocessing")

	for _, token := range path[1:] {
		tokenKey := canonicalToken(token)
		switch tokenKey {
		case "tools":
			canonical = append(canonical, "tools")
		case "presets", "presets_reproducibility":
			canonical = append(canonical, "presets")
		default:
			if toolID, ok := canonicalToolID(tokenKey); ok {
				if len(canonical) == 1 || canonical[len(canonical)-1] != "tools" {
					canonical = append(canonical, "tools")
				}
				canonical = append(canonical, toolID)
				continue
			}
			if listCommand, ok := canonicalToolListCommand(tokenKey); ok {
				if len(canonical) == 1 || canonical[len(canonical)-1] != "tools" {
					canonical = append(canonical, "tools")
				}
				canonical = append(canonical, listCommand)
				continue
			}
			if presetID, ok := canonicalPresetID(tokenKey); ok {
				if len(canonical) == 1 || canonical[len(canonical)-1] != "presets" {
					canonical = append(canonical, "presets")
				}
				canonical = append(canonical, presetID)
				continue
			}
			canonical = append(canonical, token)
		}
	}

	return canonical, true
}

func StripHelpFlags(path []string) ([]string, bool) {
	filtered := make([]string, 0, len(path))
	helpRequested := false
	for _, token := range path {
		switch strings.TrimSpace(token) {
		case "-h", "--help":
			helpRequested = true
		default:
			filtered = append(filtered, token)
		}
	}
	return filtered, helpRequested
}

func ContainsToken(path []string, token string) bool {
	for _, part := range path {
		if part == token {
			return true
		}
	}
	return false
}

func canonicalToken(raw string) string {
	return strings.TrimSpace(strings.ToLower(raw))
}

func canonicalToolID(token string) (string, bool) {
	for _, tool := range Tools() {
		if canonicalToken(tool.ID) == token {
			return tool.ID, true
		}
	}
	return "", false
}

func canonicalToolListCommand(token string) (string, bool) {
	for _, tool := range Tools() {
		if strings.TrimSpace(tool.ListCommand) == "" {
			continue
		}
		if canonicalToken(tool.ListCommand) == token {
			return tool.ListCommand, true
		}
	}
	return "", false
}

func canonicalPresetID(token string) (string, bool) {
	for _, preset := range Presets() {
		if canonicalToken(preset.ID) == token {
			return preset.ID, true
		}
	}
	return "", false
}
