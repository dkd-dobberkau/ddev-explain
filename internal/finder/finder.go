package finder

import (
	"errors"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

var ErrNoProjectFound = errors.New("no DDEV project found")
var ErrNoGlobalConfig = errors.New("no global DDEV config found")
var ErrNoProjectList = errors.New("no DDEV project list found")

// ProjectEntry represents a project entry in project_list.yaml
type ProjectEntry struct {
	AppRoot string `yaml:"approot"`
}

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

// FindAllProjects returns all known DDEV projects from project_list.yaml
func FindAllProjects() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	projectListPath := filepath.Join(homeDir, ".ddev", "project_list.yaml")
	if _, err := os.Stat(projectListPath); os.IsNotExist(err) {
		return nil, ErrNoProjectList
	}

	data, err := os.ReadFile(projectListPath)
	if err != nil {
		return nil, err
	}

	var projects map[string]ProjectEntry
	if err := yaml.Unmarshal(data, &projects); err != nil {
		return nil, err
	}

	var paths []string
	for _, entry := range projects {
		if entry.AppRoot != "" {
			paths = append(paths, entry.AppRoot)
		}
	}

	// Sort paths alphabetically
	sort.Strings(paths)

	return paths, nil
}
