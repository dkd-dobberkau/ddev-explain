package output

import (
	"encoding/json"

	"github.com/dkd-dobberkau/ddev-explain/internal/model"
)

type JSONFormatter struct{}

func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

func (f *JSONFormatter) Format(project *model.Project) (string, error) {
	data, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
