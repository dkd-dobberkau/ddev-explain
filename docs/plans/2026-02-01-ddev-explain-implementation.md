# ddev-explain Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Go CLI tool that summarizes DDEV project configurations with focus on development directories.

**Architecture:** Modular Go application with separate packages for DDEV config parsing, composer analysis, dev-path detection, and output formatting. Uses cobra for CLI, yaml.v3 for parsing.

**Tech Stack:** Go 1.21+, cobra, yaml.v3, fatih/color

---

## Task 1: Project Setup

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `cmd/root.go`

**Step 1: Initialize Go module**

Run:
```bash
go mod init github.com/ochorocho/ddev-explain
```

Expected: `go.mod` created

**Step 2: Create main.go**

```go
package main

import "github.com/ochorocho/ddev-explain/cmd"

func main() {
	cmd.Execute()
}
```

**Step 3: Create cmd/root.go with basic structure**

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	formatFlag   string
	allFlag      bool
	devPathsFlag bool
	verboseFlag  bool
)

var rootCmd = &cobra.Command{
	Use:   "ddev-explain",
	Short: "Summarize DDEV project configuration",
	Long:  `A CLI tool that analyzes DDEV projects and summarizes their configuration with focus on development directories.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("ddev-explain v0.1.0")
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&formatFlag, "format", "f", "text", "Output format: text, json, markdown")
	rootCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Show all known DDEV projects")
	rootCmd.Flags().BoolVar(&devPathsFlag, "dev-paths", false, "Show only development paths")
	rootCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Show additional details")
}
```

**Step 4: Add cobra dependency**

Run:
```bash
go get github.com/spf13/cobra
```

**Step 5: Verify it compiles and runs**

Run:
```bash
go build -o ddev-explain && ./ddev-explain --help
```

Expected: Help output with all flags shown

**Step 6: Commit**

```bash
git add go.mod go.sum main.go cmd/
git commit -m "feat: initial project setup with cobra CLI"
```

---

## Task 2: Project Model and DDEV Config Parser

**Files:**
- Create: `internal/model/model.go`
- Create: `internal/ddev/config.go`
- Create: `internal/ddev/config_test.go`

**Step 1: Create model structs**

```go
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
```

**Step 2: Write failing test for config parser**

```go
package ddev

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseConfig(t *testing.T) {
	// Create temp directory with test config
	tmpDir := t.TempDir()
	ddevDir := filepath.Join(tmpDir, ".ddev")
	os.MkdirAll(ddevDir, 0755)

	configContent := `name: test-project
type: typo3
php_version: "8.2"
webserver_type: nginx-fpm
database:
  type: mariadb
  version: "10.11"
nodejs_version: "20"
`
	os.WriteFile(filepath.Join(ddevDir, "config.yaml"), []byte(configContent), 0644)

	cfg, err := ParseConfig(tmpDir)
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	if cfg.Name != "test-project" {
		t.Errorf("expected name 'test-project', got '%s'", cfg.Name)
	}
	if cfg.Type != "typo3" {
		t.Errorf("expected type 'typo3', got '%s'", cfg.Type)
	}
	if cfg.PHPVersion != "8.2" {
		t.Errorf("expected PHP version '8.2', got '%s'", cfg.PHPVersion)
	}
	if cfg.Database.Type != "mariadb" {
		t.Errorf("expected database type 'mariadb', got '%s'", cfg.Database.Type)
	}
	if cfg.Database.Version != "10.11" {
		t.Errorf("expected database version '10.11', got '%s'", cfg.Database.Version)
	}
}
```

**Step 3: Run test to verify it fails**

Run:
```bash
go test ./internal/ddev/... -v
```

Expected: FAIL - ParseConfig undefined

**Step 4: Implement config parser**

```go
package ddev

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ochorocho/ddev-explain/internal/model"
	"gopkg.in/yaml.v3"
)

// DDEVConfig represents the raw .ddev/config.yaml structure
type DDEVConfig struct {
	Name          string            `yaml:"name"`
	Type          string            `yaml:"type"`
	PHPVersion    string            `yaml:"php_version"`
	WebserverType string            `yaml:"webserver_type"`
	Database      DatabaseConfig    `yaml:"database"`
	NodeJSVersion string            `yaml:"nodejs_version"`
	Hooks         map[string][]Hook `yaml:"hooks"`
	AdditionalServices []string     `yaml:"additional_services"`
}

type DatabaseConfig struct {
	Type    string `yaml:"type"`
	Version string `yaml:"version"`
}

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

	return project, nil
}
```

**Step 5: Add yaml dependency and run test**

Run:
```bash
go get gopkg.in/yaml.v3
go test ./internal/ddev/... -v
```

Expected: PASS

**Step 6: Commit**

```bash
git add internal/
git commit -m "feat: add project model and DDEV config parser"
```

---

## Task 3: Project Finder

**Files:**
- Create: `internal/finder/finder.go`
- Create: `internal/finder/finder_test.go`

**Step 1: Write failing test for upward search**

```go
package finder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindProjectUpward(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "myproject")
	ddevDir := filepath.Join(projectDir, ".ddev")
	subDir := filepath.Join(projectDir, "packages", "my-ext")

	os.MkdirAll(ddevDir, 0755)
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(ddevDir, "config.yaml"), []byte("name: test"), 0644)

	// Search from subdirectory
	found, err := FindProjectUpward(subDir)
	if err != nil {
		t.Fatalf("FindProjectUpward failed: %v", err)
	}

	if found != projectDir {
		t.Errorf("expected '%s', got '%s'", projectDir, found)
	}
}

func TestFindProjectUpward_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := FindProjectUpward(tmpDir)
	if err == nil {
		t.Error("expected error when no DDEV project found")
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/finder/... -v
```

Expected: FAIL - FindProjectUpward undefined

**Step 3: Implement finder**

```go
package finder

import (
	"errors"
	"os"
	"path/filepath"
)

var ErrNoProjectFound = errors.New("no DDEV project found")

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
		return nil, errors.New("no global DDEV config found")
	}

	// For now, return empty - will implement global config parsing later
	return []string{}, nil
}
```

**Step 4: Run test**

Run:
```bash
go test ./internal/finder/... -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/finder/
git commit -m "feat: add project finder with upward search"
```

---

## Task 4: Composer Parser

**Files:**
- Create: `internal/composer/composer.go`
- Create: `internal/composer/composer_test.go`

**Step 1: Write failing test**

```go
package composer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePathRepositories(t *testing.T) {
	tmpDir := t.TempDir()

	composerJSON := `{
	"name": "test/project",
	"repositories": [
		{"type": "path", "url": "./packages/*"},
		{"type": "path", "url": "../shared-ext"},
		{"type": "vcs", "url": "https://github.com/example/repo"}
	]
}`
	os.WriteFile(filepath.Join(tmpDir, "composer.json"), []byte(composerJSON), 0644)

	paths, err := ParsePathRepositories(tmpDir)
	if err != nil {
		t.Fatalf("ParsePathRepositories failed: %v", err)
	}

	if len(paths) != 2 {
		t.Errorf("expected 2 path repositories, got %d", len(paths))
	}

	// Check that paths are returned (relative)
	found := false
	for _, p := range paths {
		if p == "./packages/*" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find './packages/*' in paths")
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/composer/... -v
```

Expected: FAIL - ParsePathRepositories undefined

**Step 3: Implement composer parser**

```go
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
```

**Step 4: Run test**

Run:
```bash
go test ./internal/composer/... -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/composer/
git commit -m "feat: add composer.json parser for path repositories"
```

---

## Task 5: Dev-Path Detector

**Files:**
- Create: `internal/detector/detector.go`
- Create: `internal/detector/detector_test.go`

**Step 1: Write failing test**

```go
package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectDevPaths(t *testing.T) {
	tmpDir := t.TempDir()

	// Create packages directory with a package
	packagesDir := filepath.Join(tmpDir, "packages", "my-ext")
	os.MkdirAll(packagesDir, 0755)
	os.WriteFile(filepath.Join(packagesDir, "composer.json"), []byte(`{"type": "typo3-cms-extension"}`), 0644)

	// Create composer.json with path repo
	composerJSON := `{"repositories": [{"type": "path", "url": "./packages/*"}]}`
	os.WriteFile(filepath.Join(tmpDir, "composer.json"), []byte(composerJSON), 0644)

	paths, err := DetectDevPaths(tmpDir)
	if err != nil {
		t.Fatalf("DetectDevPaths failed: %v", err)
	}

	if len(paths) == 0 {
		t.Error("expected at least one dev path")
	}

	// Should find the packages directory
	found := false
	for _, p := range paths {
		if p.Type == "composer-path" || p.Type == "convention" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find composer-path or convention type")
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/detector/... -v
```

Expected: FAIL - DetectDevPaths undefined

**Step 3: Implement detector**

```go
package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ochorocho/ddev-explain/internal/composer"
	"github.com/ochorocho/ddev-explain/internal/model"
)

var conventionalDirs = []string{
	"packages",
	"local",
	"local-packages",
	"typo3conf/ext",
}

// DetectDevPaths finds all development directories in a project
func DetectDevPaths(projectPath string) ([]model.DevPath, error) {
	var devPaths []model.DevPath

	// 1. Composer path repositories
	composerPaths, err := detectComposerPaths(projectPath)
	if err == nil {
		devPaths = append(devPaths, composerPaths...)
	}

	// 2. Conventional directories
	conventionPaths := detectConventionalPaths(projectPath)
	devPaths = append(devPaths, conventionPaths...)

	// 3. Symlinks in vendor
	symlinkPaths, err := detectSymlinks(projectPath)
	if err == nil {
		devPaths = append(devPaths, symlinkPaths...)
	}

	// Deduplicate
	devPaths = deduplicatePaths(devPaths)

	return devPaths, nil
}

func detectComposerPaths(projectPath string) ([]model.DevPath, error) {
	var devPaths []model.DevPath

	paths, err := composer.ParsePathRepositories(projectPath)
	if err != nil {
		return nil, err
	}

	for _, p := range paths {
		absPath := p
		if !filepath.IsAbs(p) {
			absPath = filepath.Join(projectPath, p)
		}

		// Handle glob patterns
		if strings.Contains(absPath, "*") {
			matches, _ := filepath.Glob(absPath)
			for _, match := range matches {
				packages := findPackagesInDir(match)
				devPaths = append(devPaths, model.DevPath{
					Path:     match,
					Type:     "composer-path",
					Source:   "composer.json",
					Packages: packages,
				})
			}
		} else {
			packages := findPackagesInDir(absPath)
			devPaths = append(devPaths, model.DevPath{
				Path:     absPath,
				Type:     "composer-path",
				Source:   "composer.json",
				Packages: packages,
			})
		}
	}

	return devPaths, nil
}

func detectConventionalPaths(projectPath string) []model.DevPath {
	var devPaths []model.DevPath

	for _, dir := range conventionalDirs {
		fullPath := filepath.Join(projectPath, dir)
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			packages := findPackagesInDir(fullPath)
			if len(packages) > 0 {
				devPaths = append(devPaths, model.DevPath{
					Path:     fullPath,
					Type:     "convention",
					Source:   "directory pattern",
					Packages: packages,
				})
			}
		}
	}

	return devPaths
}

func detectSymlinks(projectPath string) ([]model.DevPath, error) {
	var devPaths []model.DevPath
	vendorPath := filepath.Join(projectPath, "vendor")

	if _, err := os.Stat(vendorPath); os.IsNotExist(err) {
		return devPaths, nil
	}

	filepath.Walk(vendorPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return nil
			}

			absTarget := target
			if !filepath.IsAbs(target) {
				absTarget = filepath.Join(filepath.Dir(path), target)
			}
			absTarget, _ = filepath.Abs(absTarget)

			// Check if target is outside vendor
			if !strings.HasPrefix(absTarget, vendorPath) {
				relPath, _ := filepath.Rel(projectPath, path)
				devPaths = append(devPaths, model.DevPath{
					Path:   absTarget,
					Type:   "symlink",
					Source: relPath,
				})
			}
		}
		return nil
	})

	return devPaths, nil
}

func findPackagesInDir(dir string) []string {
	var packages []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return packages
	}

	for _, entry := range entries {
		if entry.IsDir() {
			composerPath := filepath.Join(dir, entry.Name(), "composer.json")
			if _, err := os.Stat(composerPath); err == nil {
				packages = append(packages, entry.Name())
			}
		}
	}

	// Also check if dir itself is a package
	if _, err := os.Stat(filepath.Join(dir, "composer.json")); err == nil && len(packages) == 0 {
		packages = append(packages, filepath.Base(dir))
	}

	return packages
}

func deduplicatePaths(paths []model.DevPath) []model.DevPath {
	seen := make(map[string]bool)
	var result []model.DevPath

	for _, p := range paths {
		if !seen[p.Path] {
			seen[p.Path] = true
			result = append(result, p)
		}
	}

	return result
}
```

**Step 4: Run test**

Run:
```bash
go test ./internal/detector/... -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/detector/
git commit -m "feat: add dev-path detector for packages, symlinks, conventions"
```

---

## Task 6: Text Output Formatter

**Files:**
- Create: `internal/output/text.go`
- Create: `internal/output/formatter.go`

**Step 1: Create formatter interface**

```go
package output

import "github.com/ochorocho/ddev-explain/internal/model"

// Formatter defines the interface for output formatters
type Formatter interface {
	Format(project *model.Project) (string, error)
}
```

**Step 2: Implement text formatter**

```go
package output

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/ochorocho/ddev-explain/internal/model"
)

type TextFormatter struct {
	Verbose bool
}

func NewTextFormatter(verbose bool) *TextFormatter {
	return &TextFormatter{Verbose: verbose}
}

func (f *TextFormatter) Format(project *model.Project) (string, error) {
	var sb strings.Builder

	// Header
	title := color.New(color.FgCyan, color.Bold)
	label := color.New(color.FgYellow)
	value := color.New(color.FgWhite)

	sb.WriteString(title.Sprintf("DDEV Project: %s\n", project.Name))
	sb.WriteString(strings.Repeat("─", 50) + "\n\n")

	// Basic info
	sb.WriteString(label.Sprint("Type:       "))
	sb.WriteString(value.Sprintf("%s\n", project.Type))

	sb.WriteString(label.Sprint("Path:       "))
	sb.WriteString(value.Sprintf("%s\n", project.Path))

	sb.WriteString(label.Sprint("PHP:        "))
	sb.WriteString(value.Sprintf("%s\n", project.PHPVersion))

	sb.WriteString(label.Sprint("Webserver:  "))
	sb.WriteString(value.Sprintf("%s\n", project.Webserver))

	sb.WriteString(label.Sprint("Database:   "))
	sb.WriteString(value.Sprintf("%s %s\n", project.Database.Type, project.Database.Version))

	if project.NodeJS != "" {
		sb.WriteString(label.Sprint("Node.js:    "))
		sb.WriteString(value.Sprintf("%s\n", project.NodeJS))
	}

	// Development Paths
	if len(project.DevPaths) > 0 {
		sb.WriteString("\n")
		sb.WriteString(title.Sprint("📁 Development Paths\n"))
		sb.WriteString(strings.Repeat("─", 50) + "\n")

		for _, dp := range project.DevPaths {
			typeIcon := getTypeIcon(dp.Type)
			sb.WriteString(fmt.Sprintf("%s %s\n", typeIcon, dp.Path))
			sb.WriteString(fmt.Sprintf("   Type: %s | Source: %s\n", dp.Type, dp.Source))

			if len(dp.Packages) > 0 {
				sb.WriteString("   Packages: " + strings.Join(dp.Packages, ", ") + "\n")
			}
		}
	}

	// Services
	if len(project.Services) > 0 {
		sb.WriteString("\n")
		sb.WriteString(title.Sprint("🔌 Services\n"))
		sb.WriteString(strings.Repeat("─", 50) + "\n")

		for _, svc := range project.Services {
			sb.WriteString(fmt.Sprintf("• %s (%s)\n", svc.Name, svc.Type))
		}
	}

	// Commands (verbose only)
	if f.Verbose && len(project.Commands) > 0 {
		sb.WriteString("\n")
		sb.WriteString(title.Sprint("⚡ Custom Commands\n"))
		sb.WriteString(strings.Repeat("─", 50) + "\n")

		for _, cmd := range project.Commands {
			sb.WriteString(fmt.Sprintf("• %s - %s\n", cmd.Name, cmd.Description))
		}
	}

	// Hooks (verbose only)
	if f.Verbose && len(project.Hooks) > 0 {
		sb.WriteString("\n")
		sb.WriteString(title.Sprint("🪝 Hooks\n"))
		sb.WriteString(strings.Repeat("─", 50) + "\n")

		for name, cmds := range project.Hooks {
			sb.WriteString(fmt.Sprintf("• %s:\n", name))
			for _, cmd := range cmds {
				sb.WriteString(fmt.Sprintf("    - %s\n", cmd))
			}
		}
	}

	return sb.String(), nil
}

func getTypeIcon(t string) string {
	switch t {
	case "composer-path":
		return "📦"
	case "symlink":
		return "🔗"
	case "mount":
		return "💾"
	case "convention":
		return "📂"
	default:
		return "•"
	}
}
```

**Step 3: Add color dependency**

Run:
```bash
go get github.com/fatih/color
```

**Step 4: Commit**

```bash
git add internal/output/
git commit -m "feat: add text output formatter with colors"
```

---

## Task 7: JSON and Markdown Formatters

**Files:**
- Create: `internal/output/json.go`
- Create: `internal/output/markdown.go`

**Step 1: Implement JSON formatter**

```go
package output

import (
	"encoding/json"

	"github.com/ochorocho/ddev-explain/internal/model"
)

type JSONFormatter struct{}

func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

func (f *JSONFormatter) Format(project *model.Project) (string, error) {
	data, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
```

**Step 2: Implement Markdown formatter**

```go
package output

import (
	"fmt"
	"strings"

	"github.com/ochorocho/ddev-explain/internal/model"
)

type MarkdownFormatter struct {
	Verbose bool
}

func NewMarkdownFormatter(verbose bool) *MarkdownFormatter {
	return &MarkdownFormatter{Verbose: verbose}
}

func (f *MarkdownFormatter) Format(project *model.Project) (string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# DDEV Project: %s\n\n", project.Name))

	sb.WriteString("## Configuration\n\n")
	sb.WriteString("| Setting | Value |\n")
	sb.WriteString("|---------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Type | %s |\n", project.Type))
	sb.WriteString(fmt.Sprintf("| Path | `%s` |\n", project.Path))
	sb.WriteString(fmt.Sprintf("| PHP | %s |\n", project.PHPVersion))
	sb.WriteString(fmt.Sprintf("| Webserver | %s |\n", project.Webserver))
	sb.WriteString(fmt.Sprintf("| Database | %s %s |\n", project.Database.Type, project.Database.Version))
	if project.NodeJS != "" {
		sb.WriteString(fmt.Sprintf("| Node.js | %s |\n", project.NodeJS))
	}

	if len(project.DevPaths) > 0 {
		sb.WriteString("\n## Development Paths\n\n")
		for _, dp := range project.DevPaths {
			sb.WriteString(fmt.Sprintf("### %s\n\n", dp.Path))
			sb.WriteString(fmt.Sprintf("- **Type:** %s\n", dp.Type))
			sb.WriteString(fmt.Sprintf("- **Source:** %s\n", dp.Source))
			if len(dp.Packages) > 0 {
				sb.WriteString(fmt.Sprintf("- **Packages:** %s\n", strings.Join(dp.Packages, ", ")))
			}
			sb.WriteString("\n")
		}
	}

	if len(project.Services) > 0 {
		sb.WriteString("## Services\n\n")
		for _, svc := range project.Services {
			sb.WriteString(fmt.Sprintf("- **%s** (%s)\n", svc.Name, svc.Type))
		}
		sb.WriteString("\n")
	}

	if f.Verbose && len(project.Commands) > 0 {
		sb.WriteString("## Custom Commands\n\n")
		for _, cmd := range project.Commands {
			sb.WriteString(fmt.Sprintf("- `%s` - %s\n", cmd.Name, cmd.Description))
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}
```

**Step 3: Commit**

```bash
git add internal/output/json.go internal/output/markdown.go
git commit -m "feat: add JSON and Markdown output formatters"
```

---

## Task 8: CLI Integration

**Files:**
- Modify: `cmd/root.go`

**Step 1: Update root.go with full implementation**

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/ochorocho/ddev-explain/internal/ddev"
	"github.com/ochorocho/ddev-explain/internal/detector"
	"github.com/ochorocho/ddev-explain/internal/finder"
	"github.com/ochorocho/ddev-explain/internal/output"
	"github.com/spf13/cobra"
)

var (
	formatFlag     string
	allFlag        bool
	devPathsFlag   bool
	verboseFlag    bool
	installCmdFlag bool
)

var rootCmd = &cobra.Command{
	Use:   "ddev-explain",
	Short: "Summarize DDEV project configuration",
	Long:  `A CLI tool that analyzes DDEV projects and summarizes their configuration with focus on development directories.`,
	RunE:  runExplain,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&formatFlag, "format", "f", "text", "Output format: text, json, markdown")
	rootCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Show all known DDEV projects")
	rootCmd.Flags().BoolVar(&devPathsFlag, "dev-paths", false, "Show only development paths")
	rootCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Show additional details")
	rootCmd.Flags().BoolVar(&installCmdFlag, "install-command", false, "Install as DDEV custom command")
}

func runExplain(cmd *cobra.Command, args []string) error {
	if installCmdFlag {
		return installDDEVCommand()
	}

	var projectPaths []string

	if allFlag {
		paths, err := finder.FindAllProjects()
		if err != nil {
			return fmt.Errorf("failed to find projects: %w", err)
		}
		projectPaths = paths
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		projectPath, err := finder.FindProjectUpward(cwd)
		if err != nil {
			return fmt.Errorf("no DDEV project found in %s or parent directories", cwd)
		}
		projectPaths = []string{projectPath}
	}

	for _, projectPath := range projectPaths {
		project, err := ddev.ParseConfig(projectPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", projectPath, err)
			continue
		}

		// Detect dev paths
		devPaths, err := detector.DetectDevPaths(projectPath)
		if err == nil {
			project.DevPaths = devPaths
		}

		// Format output
		var formatter output.Formatter
		switch formatFlag {
		case "json":
			formatter = output.NewJSONFormatter()
		case "markdown":
			formatter = output.NewMarkdownFormatter(verboseFlag)
		default:
			formatter = output.NewTextFormatter(verboseFlag)
		}

		out, err := formatter.Format(project)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}

		fmt.Println(out)
	}

	return nil
}

func installDDEVCommand() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	cmdDir := fmt.Sprintf("%s/.ddev/commands/host", homeDir)
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		return err
	}

	cmdPath := fmt.Sprintf("%s/explain", cmdDir)
	cmdContent := `#!/bin/bash
## Description: Summarize DDEV project configuration
## Usage: explain [flags]
## Example: ddev explain --format=json

ddev-explain "$@"
`

	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0755); err != nil {
		return err
	}

	fmt.Printf("DDEV command installed: %s\n", cmdPath)
	fmt.Println("You can now use: ddev explain")
	return nil
}
```

**Step 2: Build and test**

Run:
```bash
go build -o ddev-explain
./ddev-explain --help
```

Expected: Full help output with all options

**Step 3: Commit**

```bash
git add cmd/root.go
git commit -m "feat: integrate all components in CLI"
```

---

## Task 9: Docker Mount Detection

**Files:**
- Create: `internal/detector/mounts.go`
- Modify: `internal/detector/detector.go`

**Step 1: Implement mount detection**

```go
package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ochorocho/ddev-explain/internal/model"
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
```

**Step 2: Update detector.go to include mounts**

Add to `DetectDevPaths` function after symlinks detection:

```go
	// 4. Docker mounts
	mountPaths, err := detectMounts(projectPath)
	if err == nil {
		devPaths = append(devPaths, mountPaths...)
	}
```

**Step 3: Run tests**

Run:
```bash
go test ./internal/... -v
```

Expected: All tests pass

**Step 4: Commit**

```bash
git add internal/detector/
git commit -m "feat: add Docker mount detection for external paths"
```

---

## Task 10: Service Detection

**Files:**
- Create: `internal/ddev/services.go`
- Modify: `internal/ddev/config.go`

**Step 1: Implement service detection**

```go
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
```

**Step 2: Update ParseConfig to include services**

Add at the end of `ParseConfig` before return:

```go
	// Detect services
	services, err := DetectServices(projectPath)
	if err == nil {
		project.Services = services
	}
```

**Step 3: Run tests and build**

Run:
```bash
go test ./internal/... -v
go build -o ddev-explain
```

Expected: All pass, binary builds

**Step 4: Commit**

```bash
git add internal/ddev/
git commit -m "feat: add service detection from docker-compose files"
```

---

## Task 11: Custom Commands Detection

**Files:**
- Create: `internal/ddev/commands.go`

**Step 1: Implement command detection**

```go
package ddev

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/ochorocho/ddev-explain/internal/model"
)

// DetectCommands finds custom DDEV commands in .ddev/commands/
func DetectCommands(projectPath string) ([]model.Command, error) {
	var commands []model.Command

	commandsDir := filepath.Join(projectPath, ".ddev", "commands")
	if _, err := os.Stat(commandsDir); os.IsNotExist(err) {
		return commands, nil
	}

	// Walk through host and web subdirectories
	for _, subdir := range []string{"host", "web"} {
		dir := filepath.Join(commandsDir, subdir)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			cmdPath := filepath.Join(dir, entry.Name())
			description := parseCommandDescription(cmdPath)

			commands = append(commands, model.Command{
				Name:        entry.Name(),
				Description: description,
				Path:        cmdPath,
			})
		}
	}

	return commands, nil
}

func parseCommandDescription(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "## Description:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "## Description:"))
		}
	}

	return ""
}
```

**Step 2: Update ParseConfig to include commands**

Add to `ParseConfig`:

```go
	// Detect commands
	commands, err := DetectCommands(projectPath)
	if err == nil {
		project.Commands = commands
	}
```

**Step 3: Build and test with real project**

Run:
```bash
go build -o ddev-explain
# Test with an actual DDEV project
cd /path/to/ddev/project && /path/to/ddev-explain
```

**Step 4: Commit**

```bash
git add internal/ddev/commands.go
git commit -m "feat: add custom command detection"
```

---

## Task 12: Final Testing and Polish

**Files:**
- Update: `README.md` (create)

**Step 1: Create README**

```markdown
# ddev-explain

A CLI tool that summarizes DDEV project configurations with focus on development directories.

## Installation

```bash
go install github.com/ochorocho/ddev-explain@latest
```

## Usage

```bash
# In a DDEV project directory
ddev-explain

# Show all projects
ddev-explain --all

# Different output formats
ddev-explain --format=json
ddev-explain --format=markdown

# Show only development paths
ddev-explain --dev-paths

# Verbose output (includes hooks, commands)
ddev-explain -v
```

## Install as DDEV Command

```bash
ddev-explain --install-command
ddev explain
```

## Features

- Parses DDEV configuration
- Detects development directories:
  - Composer path repositories
  - Symlinks in vendor/
  - Conventional directories (packages/, local/)
  - Docker mounts
- Lists additional services
- Shows custom commands and hooks
- Multiple output formats (text, JSON, Markdown)
```

**Step 2: Run full test suite**

Run:
```bash
go test ./... -v
go build -o ddev-explain
./ddev-explain --help
```

**Step 3: Commit**

```bash
git add README.md
git commit -m "docs: add README with usage instructions"
```

---

## Summary

The implementation is split into 12 tasks following TDD principles:

1. Project setup with cobra CLI
2. Model and DDEV config parser
3. Project finder (upward search)
4. Composer parser
5. Dev-path detector
6. Text formatter
7. JSON/Markdown formatters
8. CLI integration
9. Docker mount detection
10. Service detection
11. Command detection
12. Documentation and polish

Each task has specific files, tests, and commits. Total estimated commits: 12.
