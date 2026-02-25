package commandpalette

import (
	"strings"
)

const minExamplePoolSize = 10

func suggestionsFromCatalog(entries []Entry) []string {
	seen := map[string]struct{}{}
	suggestions := make([]string, 0, len(entries)*2)
	for _, entry := range entries {
		for _, alias := range entry.Aliases {
			value := strings.TrimSpace(alias)
			if value == "" {
				continue
			}
			if _, exists := seen[value]; exists {
				continue
			}
			seen[value] = struct{}{}
			suggestions = append(suggestions, value)
		}
	}
	return suggestions
}

func examplePool(entries []Entry) []string {
	toolEntries := filterToolEntries(entries)
	if len(toolEntries) == 0 {
		toolEntries = entries
	}

	preferred := []string{
		"data conversion",
		"synthetic data generation",
		"build trajectories",
		"combine dataset with trajectories",
		"clean dataset",
		"missing imputation",
		"harmonise categories",
		"show table",
		"aggregate long to cross sectional",
		"rename feature",
		"pivot long to wide",
		"trajectories viz",
	}
	fallback := []string{
		"benchmark standard ml",
		"benchmark longitudinal ml",
		"standard machine learning",
		"longitudinal machine learning",
		"shap analysis",
		"remove columns",
	}

	available := map[string]struct{}{}
	for _, entry := range toolEntries {
		for _, alias := range entry.Aliases {
			available[normalize(alias)] = struct{}{}
		}
	}

	examples := make([]string, 0, minExamplePoolSize)
	addExample := func(value string) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return
		}
		if containsExample(examples, trimmed) {
			return
		}
		examples = append(examples, strings.ToLower(trimmed))
	}

	for _, candidate := range preferred {
		if _, ok := available[normalize(candidate)]; ok {
			addExample(candidate)
		}
	}

	for _, entry := range toolEntries {
		addExample(entry.Label)
		if len(examples) >= minExamplePoolSize {
			return examples
		}
	}

	for _, entry := range toolEntries {
		for _, alias := range entry.Aliases {
			normalizedAlias := strings.ReplaceAll(strings.TrimSpace(alias), "_", " ")
			if len(strings.Fields(normalizedAlias)) < 2 {
				continue
			}
			addExample(alias)
			if len(examples) >= minExamplePoolSize {
				return examples
			}
		}
	}

	for _, candidate := range fallback {
		addExample(candidate)
		if len(examples) >= minExamplePoolSize {
			return examples
		}
	}

	if len(examples) == 0 {
		addExample("data conversion")
	}
	return examples
}

func filterToolEntries(entries []Entry) []Entry {
	tools := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		if len(entry.Command) < 3 {
			continue
		}
		if normalize(entry.Command[1]) != "tools" {
			continue
		}
		tools = append(tools, entry)
	}
	return tools
}

func containsExample(pool []string, value string) bool {
	for _, item := range pool {
		if normalize(item) == normalize(value) {
			return true
		}
	}
	return false
}
