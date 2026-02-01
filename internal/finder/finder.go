package finder

import (
	"errors"
	"os"
	"path/filepath"
)

var ErrNoProjectFound = errors.New("no DDEV project found")
var ErrNoGlobalConfig = errors.New("no global DDEV config found")

// FindProjectUpward searches for a DDEV project starting from startPath
// and walking up the directory tree
func FindProjectUpward(startPath string) (string, error) {
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}

	current := absPath
	for {
		configPath := filepath.Join(current, ".ddev", "config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			// Reached root
			return "", ErrNoProjectFound
		}
		current = parent
	}
}

// FindAllProjects returns all known DDEV projects from global config
func FindAllProjects() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	globalConfigPath := filepath.Join(homeDir, ".ddev", "global_config.yaml")
	if _, err := os.Stat(globalConfigPath); os.IsNotExist(err) {
		return nil, ErrNoGlobalConfig
	}

	// For now, return empty - will implement global config parsing later
	return []string{}, nil
}
