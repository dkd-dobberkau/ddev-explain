package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dkd-dobberkau/ddev-explain/internal/composer"
	"github.com/dkd-dobberkau/ddev-explain/internal/model"
)

var conventionalDirs = []string{
	"packages",
	"local",
	"local-packages",
	"typo3conf/ext",
}

// Files to ignore when detecting packages
var ignoredFiles = map[string]bool{
	".DS_Store":  true,
	".gitignore": true,
	".gitkeep":   true,
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

	// 4. Docker mounts
	mountPaths, err := detectMounts(projectPath)
	if err == nil {
		devPaths = append(devPaths, mountPaths...)
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
		// Skip container-only paths (not accessible on host)
		if strings.HasPrefix(p, "/var/www/") {
			continue
		}

		absPath := p
		if !filepath.IsAbs(p) {
			absPath = filepath.Join(projectPath, p)
		}

		// Skip ignored files
		if ignoredFiles[filepath.Base(absPath)] {
			continue
		}

		// Handle glob patterns
		if strings.Contains(absPath, "*") {
			matches, _ := filepath.Glob(absPath)
			for _, match := range matches {
				// Skip ignored files in glob matches
				if ignoredFiles[filepath.Base(match)] {
					continue
				}
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

			// Skip container-only paths
			if strings.HasPrefix(absTarget, "/var/www/") {
				return nil
			}

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

// typePriority returns priority for deduplication (lower = higher priority)
func typePriority(t string) int {
	switch t {
	case "composer-path":
		return 1
	case "symlink":
		return 2
	case "mount":
		return 3
	case "convention":
		return 4
	default:
		return 5
	}
}

func deduplicatePaths(paths []model.DevPath) []model.DevPath {
	seen := make(map[string]model.DevPath)

	for _, p := range paths {
		existing, exists := seen[p.Path]
		if !exists {
			seen[p.Path] = p
		} else if typePriority(p.Type) < typePriority(existing.Type) {
			// Keep the one with higher priority (lower number)
			seen[p.Path] = p
		}
	}

	// Also remove convention entries if their parent directory content
	// is already covered by a more specific type
	parentCovered := make(map[string]bool)
	for path, dp := range seen {
		if dp.Type != "convention" {
			parentCovered[filepath.Dir(path)] = true
		}
	}

	var result []model.DevPath
	for _, dp := range seen {
		// Skip convention entries if a child path is already covered
		if dp.Type == "convention" && parentCovered[dp.Path] {
			continue
		}
		result = append(result, dp)
	}

	return result
}
