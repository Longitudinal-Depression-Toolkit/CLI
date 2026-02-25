package common

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func FindProjectRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := currentDir
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "pyproject.toml")); statErr == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", errors.New("could not locate project root (missing pyproject.toml)")
}

func ResolveConfigPath(candidate string) (string, error) {
	cleanCandidate := filepath.Clean(strings.TrimSpace(candidate))
	if cleanCandidate == "" {
		return "", errors.New("empty config path")
	}
	if filepath.IsAbs(cleanCandidate) {
		return cleanCandidate, nil
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(currentDir, cleanCandidate), nil
}
