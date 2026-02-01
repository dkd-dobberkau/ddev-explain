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
	if err := os.MkdirAll(packagesDir, 0755); err != nil {
		t.Fatalf("failed to create packages dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packagesDir, "composer.json"), []byte(`{"type": "typo3-cms-extension"}`), 0644); err != nil {
		t.Fatalf("failed to write composer.json: %v", err)
	}

	// Create composer.json with path repo
	composerJSON := `{"repositories": [{"type": "path", "url": "./packages/*"}]}`
	if err := os.WriteFile(filepath.Join(tmpDir, "composer.json"), []byte(composerJSON), 0644); err != nil {
		t.Fatalf("failed to write project composer.json: %v", err)
	}

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

func TestDetectDevPaths_ConventionalDirs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create local-packages directory with a package (no composer.json in root)
	localPkgDir := filepath.Join(tmpDir, "local-packages", "test-package")
	if err := os.MkdirAll(localPkgDir, 0755); err != nil {
		t.Fatalf("failed to create local-packages dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localPkgDir, "composer.json"), []byte(`{"name": "test/package"}`), 0644); err != nil {
		t.Fatalf("failed to write composer.json: %v", err)
	}

	paths, err := DetectDevPaths(tmpDir)
	if err != nil {
		t.Fatalf("DetectDevPaths failed: %v", err)
	}

	// Should find conventional directory
	found := false
	for _, p := range paths {
		if p.Type == "convention" && len(p.Packages) > 0 {
			found = true
			if p.Packages[0] != "test-package" {
				t.Errorf("expected package 'test-package', got '%s'", p.Packages[0])
			}
			break
		}
	}
	if !found {
		t.Error("expected to find convention type with packages")
	}
}

func TestDetectDevPaths_EmptyProject(t *testing.T) {
	tmpDir := t.TempDir()

	paths, err := DetectDevPaths(tmpDir)
	if err != nil {
		t.Fatalf("DetectDevPaths failed: %v", err)
	}

	if len(paths) != 0 {
		t.Errorf("expected no dev paths for empty project, got %d", len(paths))
	}
}

func TestDetectDevPaths_Symlinks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a package outside vendor
	extDir := filepath.Join(tmpDir, "my-extensions", "ext1")
	if err := os.MkdirAll(extDir, 0755); err != nil {
		t.Fatalf("failed to create ext dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(extDir, "composer.json"), []byte(`{}`), 0644); err != nil {
		t.Fatalf("failed to write composer.json: %v", err)
	}

	// Create vendor directory with a symlink
	vendorDir := filepath.Join(tmpDir, "vendor", "my-vendor", "ext1")
	if err := os.MkdirAll(filepath.Dir(vendorDir), 0755); err != nil {
		t.Fatalf("failed to create vendor dir: %v", err)
	}
	if err := os.Symlink(extDir, vendorDir); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	paths, err := DetectDevPaths(tmpDir)
	if err != nil {
		t.Fatalf("DetectDevPaths failed: %v", err)
	}

	// Should find symlink
	found := false
	for _, p := range paths {
		if p.Type == "symlink" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find symlink type")
	}
}

func TestDetectDevPaths_Deduplication(t *testing.T) {
	tmpDir := t.TempDir()

	// Create packages directory with a package
	packagesDir := filepath.Join(tmpDir, "packages", "my-ext")
	if err := os.MkdirAll(packagesDir, 0755); err != nil {
		t.Fatalf("failed to create packages dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packagesDir, "composer.json"), []byte(`{}`), 0644); err != nil {
		t.Fatalf("failed to write composer.json: %v", err)
	}

	// Create composer.json with path repo pointing to same directory
	composerJSON := `{"repositories": [{"type": "path", "url": "./packages/*"}]}`
	if err := os.WriteFile(filepath.Join(tmpDir, "composer.json"), []byte(composerJSON), 0644); err != nil {
		t.Fatalf("failed to write project composer.json: %v", err)
	}

	paths, err := DetectDevPaths(tmpDir)
	if err != nil {
		t.Fatalf("DetectDevPaths failed: %v", err)
	}

	// Count how many times packages/my-ext appears
	count := 0
	for _, p := range paths {
		if filepath.Base(p.Path) == "my-ext" {
			count++
		}
	}

	if count > 1 {
		t.Errorf("expected deduplication, but found %d entries for my-ext", count)
	}
}
