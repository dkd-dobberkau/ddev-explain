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
	sb.WriteString(strings.Repeat("-", 50) + "\n\n")

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
		sb.WriteString(title.Sprint("Development Paths\n"))
		sb.WriteString(strings.Repeat("-", 50) + "\n")

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
		sb.WriteString(title.Sprint("Services\n"))
		sb.WriteString(strings.Repeat("-", 50) + "\n")

		for _, svc := range project.Services {
			sb.WriteString(fmt.Sprintf("* %s (%s)\n", svc.Name, svc.Type))
		}
	}

	// Commands (verbose only)
	if f.Verbose && len(project.Commands) > 0 {
		sb.WriteString("\n")
		sb.WriteString(title.Sprint("Custom Commands\n"))
		sb.WriteString(strings.Repeat("-", 50) + "\n")

		for _, cmd := range project.Commands {
			sb.WriteString(fmt.Sprintf("* %s - %s\n", cmd.Name, cmd.Description))
		}
	}

	// Hooks (verbose only)
	if f.Verbose && len(project.Hooks) > 0 {
		sb.WriteString("\n")
		sb.WriteString(title.Sprint("Hooks\n"))
		sb.WriteString(strings.Repeat("-", 50) + "\n")

		for name, cmds := range project.Hooks {
			sb.WriteString(fmt.Sprintf("* %s:\n", name))
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
		return "[pkg]"
	case "symlink":
		return "[lnk]"
	case "mount":
		return "[mnt]"
	case "convention":
		return "[dir]"
	default:
		return "*"
	}
}
