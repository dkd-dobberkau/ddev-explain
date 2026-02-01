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
