package ddev

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ochorocho/ddev-explain/internal/model"
	"gopkg.in/yaml.v3"
)

// DetectServices finds additional services from docker-compose files
func DetectServices(projectPath string) ([]model.Service, error) {
	var services []model.Service

	ddevDir := filepath.Join(projectPath, ".ddev")
	entries, err := os.ReadDir(ddevDir)
	if err != nil {
		return services, err
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "docker-compose.") &&
			strings.HasSuffix(entry.Name(), ".yaml") &&
			entry.Name() != "docker-compose.yaml" {

			filePath := filepath.Join(ddevDir, entry.Name())
			svcList, err := parseServicesFromCompose(filePath)
			if err == nil {
				services = append(services, svcList...)
			}
		}
	}

	return services, nil
}

func parseServicesFromCompose(filePath string) ([]model.Service, error) {
	var services []model.Service

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var dc struct {
		Services map[string]struct {
			Image string   `yaml:"image"`
			Ports []string `yaml:"ports"`
		} `yaml:"services"`
	}

	if err := yaml.Unmarshal(data, &dc); err != nil {
		return nil, err
	}

	for name, svc := range dc.Services {
		// Skip standard DDEV services
		if name == "web" || name == "db" || name == "dba" {
			continue
		}

		serviceType := "custom"
		if strings.Contains(svc.Image, "solr") {
			serviceType = "solr"
		} else if strings.Contains(svc.Image, "redis") {
			serviceType = "redis"
		} else if strings.Contains(svc.Image, "elastic") {
			serviceType = "elasticsearch"
		} else if strings.Contains(svc.Image, "mailhog") || strings.Contains(svc.Image, "mailpit") {
			serviceType = "mail"
		}

		services = append(services, model.Service{
			Name:  name,
			Type:  serviceType,
			Ports: svc.Ports,
		})
	}

	return services, nil
}
