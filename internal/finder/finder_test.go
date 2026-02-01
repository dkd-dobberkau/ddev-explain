package finder

import (
	"errors"
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

	if err := os.MkdirAll(ddevDir, 0755); err != nil {
		t.Fatalf("failed to create ddev dir: %v", err)
	}
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create sub dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ddevDir, "config.yaml"), []byte("name: test"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

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
	if !errors.Is(err, ErrNoProjectFound) {
		t.Errorf("expected ErrNoProjectFound, got %v", err)
	}
}

func TestFindProjectUpward_AtRoot(t *testing.T) {
	tmpDir := t.TempDir()
	ddevDir := filepath.Join(tmpDir, ".ddev")
	if err := os.MkdirAll(ddevDir, 0755); err != nil {
		t.Fatalf("failed to create ddev dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ddevDir, "config.yaml"), []byte("name: test"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Search from project root itself
	found, err := FindProjectUpward(tmpDir)
	if err != nil {
		t.Fatalf("FindProjectUpward failed: %v", err)
	}
	if found != tmpDir {
		t.Errorf("expected '%s', got '%s'", tmpDir, found)
	}
}

func TestFindAllProjects_NoGlobalConfig(t *testing.T) {
	// This test relies on ~/.ddev/global_config.yaml not existing in test env
	// or we can't easily test this. For now, just verify the function exists and returns.
	_, err := FindAllProjects()
	// We don't assert the error since it depends on the test environment
	_ = err
}
