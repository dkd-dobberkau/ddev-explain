package model

// Project represents a complete DDEV project analysis
type Project struct {
	Name       string              `json:"name"`
	Path       string              `json:"path"`
	Type       string              `json:"type"`
	PHPVersion string              `json:"php_version"`
	Webserver  string              `json:"webserver"`
	Database   Database            `json:"database"`
	URLs       []string            `json:"urls,omitempty"`
	NodeJS     string              `json:"nodejs,omitempty"`
	Services   []Service           `json:"services,omitempty"`
	DevPaths   []DevPath           `json:"dev_paths,omitempty"`
	Commands   []Command           `json:"commands,omitempty"`
	Hooks      map[string][]string `json:"hooks,omitempty"`
}

// Database represents database configuration
type Database struct {
	Type    string `json:"type"`
	Version string `json:"version"`
}

// Service represents an additional DDEV service
type Service struct {
	Name   string                 `json:"name"`
	Type   string                 `json:"type"`
	Ports  []string               `json:"ports,omitempty"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// DevPath represents a development directory
type DevPath struct {
	Path        string   `json:"path"`
	Type        string   `json:"type"`          // "composer-path", "symlink", "mount", "convention"
	Source      string   `json:"source"`        // Where detected (composer.json, docker-compose, etc.)
	MountTarget string   `json:"mount_target,omitempty"` // If mount: target in container
	Packages    []string `json:"packages,omitempty"`
}

// Command represents a DDEV custom command
type Command struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Path        string `json:"path"`
}
