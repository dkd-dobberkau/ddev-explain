package composer

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ComposerJSON struct {
	Name         string       `json:"name"`
	Type         string       `json:"type"`
	Repositories []Repository `json:"repositories"`
}

type Repository struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// ParsePathRepositories extracts path-type repositories from composer.json
func ParsePathRepositories(projectPath string) ([]string, error) {
	composerPath := filepath.Join(projectPath, "composer.json")

	data, err := os.ReadFile(composerPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var composer ComposerJSON
	if err := json.Unmarshal(data, &composer); err != nil {
		return nil, err
	}

	var paths []string
	for _, repo := range composer.Repositories {
		if repo.Type == "path" {
			paths = append(paths, repo.URL)
		}
	}

	return paths, nil
}

// GetPackageType returns the type field from a composer.json
func GetPackageType(packagePath string) (string, error) {
	composerPath := filepath.Join(packagePath, "composer.json")

	data, err := os.ReadFile(composerPath)
	if err != nil {
		return "", err
	}

	var composer ComposerJSON
	if err := json.Unmarshal(data, &composer); err != nil {
		return "", err
	}

	return composer.Type, nil
}
