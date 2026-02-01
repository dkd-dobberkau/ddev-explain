package ddev

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dkd-dobberkau/ddev-explain/internal/model"
	"gopkg.in/yaml.v3"
)

// DDEVConfig represents the raw .ddev/config.yaml structure
type DDEVConfig struct {
	Name               string            `yaml:"name"`
	Type               string            `yaml:"type"`
	PHPVersion         string            `yaml:"php_version"`
	WebserverType      string            `yaml:"webserver_type"`
	Database           DatabaseConfig    `yaml:"database"`
	NodeJSVersion      string            `yaml:"nodejs_version"`
	Hooks              map[string][]Hook `yaml:"hooks"`
	AdditionalServices []string          `yaml:"additional_services"`
}

// DatabaseConfig represents the database section in config.yaml
type DatabaseConfig struct {
	Type    string `yaml:"type"`
	Version string `yaml:"version"`
}

// Hook represents a hook command in config.yaml
type Hook struct {
	Exec     string `yaml:"exec"`
	ExecHost string `yaml:"exec-host"`
}

// ParseConfig reads and parses the DDEV config from a project directory
func ParseConfig(projectPath string) (*model.Project, error) {
	configPath := filepath.Join(projectPath, ".ddev", "config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg DDEVConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	project := &model.Project{
		Name:       cfg.Name,
		Path:       projectPath,
		Type:       cfg.Type,
		PHPVersion: cfg.PHPVersion,
		Webserver:  cfg.WebserverType,
		Database: model.Database{
			Type:    cfg.Database.Type,
			Version: cfg.Database.Version,
		},
		NodeJS: cfg.NodeJSVersion,
		Hooks:  make(map[string][]string),
	}

	// Convert hooks
	for hookName, hooks := range cfg.Hooks {
		for _, h := range hooks {
			if h.Exec != "" {
				project.Hooks[hookName] = append(project.Hooks[hookName], h.Exec)
			}
			if h.ExecHost != "" {
				project.Hooks[hookName] = append(project.Hooks[hookName], "(host) "+h.ExecHost)
			}
		}
	}

	// Detect services
	services, err := DetectServices(projectPath)
	if err == nil {
		project.Services = services
	}

	// Detect commands
	commands, err := DetectCommands(projectPath)
	if err == nil {
		project.Commands = commands
	}

	return project, nil
}
