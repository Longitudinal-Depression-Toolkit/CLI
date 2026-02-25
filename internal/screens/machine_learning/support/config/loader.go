package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	toolkitconfig "ldt-toolkit-cli/config"
	"ldt-toolkit-cli/internal/shared/common"
)

func loadConfig() (Config, error) {
	merged, err := readBundledConfig()
	if err != nil {
		return Config{}, err
	}

	overridePath, hasOverride, err := findOverrideConfigPath()
	if err != nil {
		return Config{}, err
	}
	if !hasOverride {
		return merged, nil
	}

	overrideConfig, err := readConfigFile(overridePath)
	if err != nil {
		return Config{}, err
	}

	if value := strings.TrimSpace(overrideConfig.RootSummary); value != "" {
		merged.RootSummary = value
	}
	if value := strings.TrimSpace(overrideConfig.ToolsSummary); value != "" {
		merged.ToolsSummary = value
	}
	if value := strings.TrimSpace(overrideConfig.PresetsSummary); value != "" {
		merged.PresetsSummary = value
	}
	if len(overrideConfig.Tools) > 0 {
		merged.Tools = overrideConfig.Tools
	}
	if len(overrideConfig.Presets) > 0 {
		merged.Presets = overrideConfig.Presets
	}

	if len(merged.Tools) == 0 {
		return Config{}, errors.New("machine learning config produced zero tools; keep at least one tool")
	}
	return merged, nil
}

func readBundledConfig() (Config, error) {
	raw, embedErr := toolkitconfig.ReadBundledConfigFile(defaultDataMachineLearningConfigPath)
	if embedErr == nil {
		return decodeConfig(raw, defaultDataMachineLearningConfigPath)
	}

	basePath, pathErr := findBundledConfigPath()
	if pathErr != nil {
		return Config{}, fmt.Errorf(
			"failed to load bundled machine learning config, embedded error: %v, filesystem error: %w",
			embedErr,
			pathErr,
		)
	}
	return readConfigFile(basePath)
}

func findBundledConfigPath() (string, error) {
	candidates := []string{
		defaultDataMachineLearningConfigPath,
		filepath.Join("cli", defaultDataMachineLearningConfigPath),
		filepath.Join("src", "cli", defaultDataMachineLearningConfigPath),
	}

	if root, err := common.FindProjectRoot(); err == nil {
		candidates = append(
			candidates,
			filepath.Join(root, "src", "cli", defaultDataMachineLearningConfigPath),
			filepath.Join(root, "cli", defaultDataMachineLearningConfigPath),
		)
	}

	for _, candidate := range candidates {
		resolved, err := common.ResolveConfigPath(candidate)
		if err != nil {
			continue
		}
		if _, err := os.Stat(resolved); err == nil {
			return resolved, nil
		}
	}

	return "", fmt.Errorf("machine learning config file not found; expected %s", defaultDataMachineLearningConfigPath)
}

func findOverrideConfigPath() (string, bool, error) {
	if value := strings.TrimSpace(os.Getenv(dataMachineLearningConfigEnvVar)); value != "" {
		resolved, err := common.ResolveConfigPath(value)
		if err != nil {
			return "", false, err
		}
		return resolved, true, nil
	}
	return "", false, nil
}

func readConfigFile(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read machine learning config at %s: %w", path, err)
	}
	return decodeConfig(raw, path)
}

func decodeConfig(raw []byte, source string) (Config, error) {
	var decoded Config
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return Config{}, fmt.Errorf("invalid JSON in %s: %w", source, err)
	}

	decoded.Tools = sanitiseTools(decoded.Tools)
	decoded.Presets = sanitisePresets(decoded.Presets)
	if len(decoded.Tools) == 0 {
		return Config{}, errors.New("machine learning config produced zero tools; keep at least one tool")
	}
	return decoded, nil
}

func sanitiseTools(tools []ToolConfig) []ToolConfig {
	cleaned := make([]ToolConfig, 0, len(tools))
	seen := make(map[string]struct{}, len(tools))
	for _, tool := range tools {
		id := strings.TrimSpace(tool.ID)
		if id == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}

		name := strings.TrimSpace(tool.Name)
		if name == "" {
			name = "Unnamed tool"
		}
		desc := strings.TrimSpace(tool.Description)
		if desc == "" {
			desc = name
		}
		nodeSummary := strings.TrimSpace(tool.NodeSummary)
		if nodeSummary == "" {
			nodeSummary = desc
		}
		subtitle := strings.TrimSpace(tool.Subtitle)
		if subtitle == "" {
			subtitle = nodeSummary
		}
		tableTitle := strings.TrimSpace(tool.TableTitle)
		if tableTitle == "" {
			tableTitle = name
		}
		selectionTitle := strings.TrimSpace(tool.SelectionTitle)
		if selectionTitle == "" {
			selectionTitle = name + " technique"
		}
		catalogOperation := strings.TrimSpace(tool.CatalogOperation)
		runOperation := strings.TrimSpace(tool.RunOperation)
		status := strings.TrimSpace(tool.Status)
		if status == "" {
			status = "available"
		}
		if strings.EqualFold(status, "available") && (catalogOperation == "" || runOperation == "") {
			continue
		}

		cleaned = append(cleaned, ToolConfig{
			ID:               id,
			Name:             name,
			Description:      desc,
			NodeSummary:      nodeSummary,
			Subtitle:         subtitle,
			TableTitle:       tableTitle,
			SelectionTitle:   selectionTitle,
			CatalogOperation: catalogOperation,
			RunOperation:     runOperation,
			ListCommand:      strings.TrimSpace(tool.ListCommand),
			Status:           status,
		})
		seen[id] = struct{}{}
	}
	return cleaned
}

func sanitisePresets(presets []PresetConfig) []PresetConfig {
	cleaned := make([]PresetConfig, 0, len(presets))
	seen := make(map[string]struct{}, len(presets))
	for _, preset := range presets {
		id := strings.TrimSpace(preset.ID)
		if id == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}

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
			status = "incoming"
		}

		cleaned = append(cleaned, PresetConfig{
			ID:          id,
			Name:        name,
			Description: description,
			Status:      status,
		})
		seen[id] = struct{}{}
	}
	return cleaned
}
