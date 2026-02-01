package model

// Project represents a complete DDEV project analysis
type Project struct {
	Name       string
	Path       string
	Type       string
	PHPVersion string
	Webserver  string
	Database   Database
	URLs       []string
	NodeJS     string
	Services   []Service
	DevPaths   []DevPath
	Commands   []Command
	Hooks      map[string][]string
}

// Database represents database configuration
type Database struct {
	Type    string
	Version string
}

// Service represents an additional DDEV service
type Service struct {
	Name   string
	Type   string
	Ports  []string
	Config map[string]interface{}
}

// DevPath represents a development directory
type DevPath struct {
	Path        string
	Type        string // "composer-path", "symlink", "mount", "convention"
	Source      string // Where detected (composer.json, docker-compose, etc.)
	MountTarget string // If mount: target in container
	Packages    []string
}

// Command represents a DDEV custom command
type Command struct {
	Name        string
	Description string
	Path        string
}
