package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dkd-dobberkau/ddev-explain/internal/model"
	"gopkg.in/yaml.v3"
)

// DockerCompose represents relevant parts of docker-compose files
type DockerCompose struct {
	Services map[string]struct {
		Volumes []string `yaml:"volumes"`
	} `yaml:"services"`
}

func detectMounts(projectPath string) ([]model.DevPath, error) {
	var devPaths []model.DevPath

	// Find docker-compose.*.yaml files
	ddevDir := filepath.Join(projectPath, ".ddev")
	entries, err := os.ReadDir(ddevDir)
	if err != nil {
		return devPaths, err
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "docker-compose.") &&
			strings.HasSuffix(entry.Name(), ".yaml") &&
			entry.Name() != "docker-compose.yaml" {

			filePath := filepath.Join(ddevDir, entry.Name())
			mounts, err := parseDockerCompose(filePath, projectPath)
			if err == nil {
				devPaths = append(devPaths, mounts...)
			}
		}
	}

	return devPaths, nil
}

func parseDockerCompose(filePath, projectPath string) ([]model.DevPath, error) {
	var devPaths []model.DevPath

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var dc DockerCompose
	if err := yaml.Unmarshal(data, &dc); err != nil {
		return nil, err
	}

	for _, svc := range dc.Services {
		for _, vol := range svc.Volumes {
			parts := strings.SplitN(vol, ":", 2)
			if len(parts) < 2 {
				continue
			}

			hostPath := parts[0]
			containerPath := parts[1]

			// Skip non-path volumes
			if !strings.HasPrefix(hostPath, ".") && !strings.HasPrefix(hostPath, "/") {
				continue
			}

			absPath := hostPath
			if !filepath.IsAbs(hostPath) {
				absPath = filepath.Join(filepath.Dir(filePath), hostPath)
			}
			absPath, _ = filepath.Abs(absPath)

			// Only include paths outside project
			if !strings.HasPrefix(absPath, projectPath) {
				devPaths = append(devPaths, model.DevPath{
					Path:        absPath,
					Type:        "mount",
					Source:      filepath.Base(filePath),
					MountTarget: containerPath,
				})
			}
		}
	}

	return devPaths, nil
}
