package commandpalette

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const historyMaxItems = 128

type historyFile struct {
	Commands []string `json:"commands"`
}

type historyStore struct {
	path string
}

func newHistoryStore(projectRootFinder func() (string, error)) historyStore {
	path := filepath.Join(os.TempDir(), "ldt_toolkit_command_palette_history.json")
	if projectRootFinder != nil {
		if root, err := projectRootFinder(); err == nil {
			path = filepath.Join(root, "tmp", "command_palette_history.json")
		}
	}
	return historyStore{path: path}
}

func (s historyStore) Load() []string {
	raw, err := os.ReadFile(s.path)
	if err != nil {
		return nil
	}
	var payload historyFile
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil
	}
	cleaned := make([]string, 0, len(payload.Commands))
	seen := map[string]struct{}{}
	for _, command := range payload.Commands {
		value := strings.TrimSpace(command)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		cleaned = append(cleaned, value)
		if len(cleaned) >= historyMaxItems {
			break
		}
	}
	return cleaned
}

func (s historyStore) Save(commands []string) error {
	if strings.TrimSpace(s.path) == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	payload := historyFile{Commands: commands}
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, raw, 0o644)
}
