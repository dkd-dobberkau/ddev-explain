package output

import "github.com/dkd-dobberkau/ddev-explain/internal/model"

// Formatter defines the interface for output formatters
type Formatter interface {
	Format(project *model.Project) (string, error)
}
