package commandpalette

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

const (
	commandPaletteConfigEnvVar      = "LDT_COMMAND_PALETTE_CONFIG"
	defaultCommandPaletteConfigPath = "config/command_palette.json"
)

type catalogConfig struct {
	Entries []catalogConfigEntry `json:"entries"`
}

type catalogConfigEntry struct {
	Label    string   `json:"label"`
	Command  []string `json:"command"`
	Keywords []string `json:"keywords"`
}

func loadCatalogFromConfig() ([]Entry, error) {
	raw, source, err := readBundledCatalogConfig()
	if err != nil {
		return nil, err
	}

	overridePath, hasOverride, err := findCatalogOverrideConfigPath()
	if err != nil {
		return nil, err
	}
	if hasOverride {
		raw, err = os.ReadFile(overridePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read command palette config at %s: %w", overridePath, err)
		}
		source = overridePath
	}

	return parseCatalogConfig(raw, source)
}

func readBundledCatalogConfig() ([]byte, string, error) {
	raw, embedErr := toolkitconfig.ReadBundledConfigFile(defaultCommandPaletteConfigPath)
	if embedErr == nil {
		return raw, defaultCommandPaletteConfigPath, nil
	}

	configPath, hasConfig, pathErr := findBundledCatalogConfigPath()
	if pathErr != nil {
		return nil, "", fmt.Errorf(
			"failed to load bundled command palette config, embedded error: %v, filesystem error: %w",
			embedErr,
			pathErr,
		)
	}
	if !hasConfig {
		return nil, "", errors.New("command palette config file not found")
	}

	raw, err := os.ReadFile(configPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read command palette config at %s: %w", configPath, err)
	}
	return raw, configPath, nil
}

func parseCatalogConfig(raw []byte, source string) ([]Entry, error) {
	var decoded catalogConfig
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, fmt.Errorf("invalid JSON in %s: %w", source, err)
	}

	entries := make([]Entry, 0, len(decoded.Entries))
	for _, configured := range decoded.Entries {
		command := sanitiseCommand(configured.Command)
		if len(command) == 0 {
			continue
		}

		label := strings.TrimSpace(configured.Label)
		if label == "" {
			label = strings.Join(command, " ")
		}

		entries = append(entries, Entry{
			Label:   label,
			Command: command,
			Aliases: sanitiseKeywords(configured.Keywords),
		})
	}

	if len(entries) == 0 {
		return nil, errors.New("command palette config produced zero valid entries")
	}
	return entries, nil
}

func findCatalogOverrideConfigPath() (string, bool, error) {
	if value := strings.TrimSpace(os.Getenv(commandPaletteConfigEnvVar)); value != "" {
		resolved, err := common.ResolveConfigPath(value)
		if err != nil {
			return "", false, err
		}
		return resolved, true, nil
	}
	return "", false, nil
}

func findBundledCatalogConfigPath() (string, bool, error) {
	candidates := []string{
		defaultCommandPaletteConfigPath,
		filepath.Join("cli", defaultCommandPaletteConfigPath),
		filepath.Join("src", "cli", defaultCommandPaletteConfigPath),
	}

	if root, err := common.FindProjectRoot(); err == nil {
		candidates = append(
			candidates,
			filepath.Join(root, "src", "cli", defaultCommandPaletteConfigPath),
			filepath.Join(root, "cli", defaultCommandPaletteConfigPath),
		)
	}

	for _, candidate := range candidates {
		resolved, err := common.ResolveConfigPath(candidate)
		if err != nil {
			continue
		}
		if _, err := os.Stat(resolved); err == nil {
			return resolved, true, nil
		}
	}

	return "", false, nil
}

func sanitiseCommand(command []string) []string {
	cleaned := make([]string, 0, len(command))
	for _, token := range command {
		trimmed := strings.TrimSpace(token)
		if trimmed == "" {
			continue
		}
		cleaned = append(cleaned, trimmed)
	}
	return cleaned
}

func sanitiseKeywords(keywords []string) []string {
	cleaned := make([]string, 0, len(keywords))
	for _, keyword := range keywords {
		trimmed := strings.TrimSpace(keyword)
		if trimmed == "" {
			continue
		}
		cleaned = append(cleaned, trimmed)
	}
	return cleaned
}
