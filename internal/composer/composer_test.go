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
	if err := os.WriteFile(filepath.Join(tmpDir, "composer.json"), []byte(composerJSON), 0644); err != nil {
		t.Fatalf("failed to write composer.json: %v", err)
	}

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

func TestParsePathRepositories_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	paths, err := ParsePathRepositories(tmpDir)
	if err != nil {
		t.Fatalf("expected no error for missing composer.json, got %v", err)
	}
	if len(paths) != 0 {
		t.Errorf("expected empty paths for missing composer.json, got %d", len(paths))
	}
}
