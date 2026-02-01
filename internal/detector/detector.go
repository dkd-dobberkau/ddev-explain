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
