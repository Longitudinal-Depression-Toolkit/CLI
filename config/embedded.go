package toolkitconfig

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"
)

// bundledConfigFiles stores the default JSON configuration shipped with the CLI binary.
//
//go:embed *.json
var bundledConfigFiles embed.FS

// ReadBundledConfigFile reads a bundled JSON configuration file by name.
//
// The function accepts either a bare filename (e.g. "home.json") or a path
// (e.g. "config/home.json"), and always resolves to the base filename.
func ReadBundledConfigFile(name string) ([]byte, error) {
	baseName := strings.TrimSpace(filepath.Base(name))
	if baseName == "" || baseName == "." {
		return nil, fmt.Errorf("invalid bundled config name: %q", name)
	}

	raw, err := bundledConfigFiles.ReadFile(baseName)
	if err != nil {
		return nil, fmt.Errorf("bundled config file %q not found: %w", baseName, err)
	}
	return raw, nil
}
