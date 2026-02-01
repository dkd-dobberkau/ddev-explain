package output

import (
	"fmt"
	"strings"

	"github.com/dkd-dobberkau/ddev-explain/internal/model"
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
