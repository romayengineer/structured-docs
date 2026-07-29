package renderer

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/romayengineer/structured-docs/pkg/structured/resolver"
)

func Render(job *resolver.Job) (string, error) {
	tmpl, err := template.New(job.Template.FileName).Parse(job.Template.Content)
	if err != nil {
		return "", fmt.Errorf("parsing template %s: %w", job.Template.FileName, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, job.Data.Fields); err != nil {
		return "", fmt.Errorf("rendering template %s with data %s: %w", job.Template.FileName, job.Data.SourcePath, err)
	}

	return buf.String(), nil
}
