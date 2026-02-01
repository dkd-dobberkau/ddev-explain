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
	if err := os.MkdirAll(ddevDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	configContent := `name: test-project
type: typo3
php_version: "8.2"
webserver_type: nginx-fpm
database:
  type: mariadb
  version: "10.11"
nodejs_version: "20"
`
	if err := os.WriteFile(filepath.Join(ddevDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

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

func TestParseConfig_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := ParseConfig(tmpDir)
	if err == nil {
		t.Error("expected error when config file is missing")
	}
}
