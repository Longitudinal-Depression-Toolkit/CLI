package commandpalette

import (
	"errors"
	"strings"
)

func buildCatalog() ([]Entry, error) {
	entries, err := loadCatalogFromConfig()
	if err != nil {
		return nil, err
	}

	catalog := sanitizeCatalog(entries)
	if len(catalog) == 0 {
		return nil, errors.New("command palette config produced zero usable entries")
	}
	return catalog, nil
}

func sanitizeCatalog(entries []Entry) []Entry {
	cleaned := make([]Entry, 0, len(entries))
	seenCommands := make(map[string]struct{}, len(entries))

	for _, entry := range entries {
		label := strings.TrimSpace(entry.Label)
		if label == "" {
			label = "Unnamed"
		}
		command := clonePath(entry.Command)
		if len(command) == 0 {
			continue
		}
		for index := range command {
			command[index] = strings.TrimSpace(command[index])
		}
		key := strings.Join(command, "\x1f")
		if _, exists := seenCommands[key]; exists {
			continue
		}

		aliases := expandAliases(append([]string{label}, entry.Aliases...)...)
		cleaned = append(cleaned, Entry{
			Label:   label,
			Command: command,
			Aliases: aliases,
		})
		seenCommands[key] = struct{}{}
	}

	return cleaned
}

func expandAliases(values ...string) []string {
	seen := map[string]struct{}{}
	aliases := make([]string, 0, len(values)*2)

	add := func(value string) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return
		}
		if _, exists := seen[trimmed]; exists {
			return
		}
		seen[trimmed] = struct{}{}
		aliases = append(aliases, trimmed)
	}

	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		add(trimmed)
		add(strings.ReplaceAll(trimmed, "_", " "))
		add(strings.ReplaceAll(trimmed, "-", " "))

		normalized := normalize(trimmed)
		add(normalized)
		for _, token := range strings.Fields(normalized) {
			if len(token) >= 3 {
				add(token)
			}
		}
	}

	return aliases
}

func clonePath(path []string) []string {
	copied := make([]string, len(path))
	copy(copied, path)
	return copied
}
