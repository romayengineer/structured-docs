package resolver

import (
	"fmt"

	"github.com/romayengineer/structured-docs/pkg/structured/data"
	"github.com/romayengineer/structured-docs/pkg/structured/schema"
	"github.com/romayengineer/structured-docs/pkg/structured/template"
)

type Job struct {
	Data     *data.DataFile
	Template *template.Template
}

func ResolveAll(
	dataFiles []*data.DataFile,
	templates []*template.Template,
	templateOrder []string,
	types map[string]*schema.TypeDefinition,
) ([]*Job, error) {
	tmplByName := make(map[string]*template.Template)
	for _, t := range templates {
		tmplByName[t.FileName] = t
	}

	var jobs []*Job

	for _, df := range dataFiles {
		if df.ExplicitTemplate != "" {
			t, ok := tmplByName[df.ExplicitTemplate]
			if !ok {
				return nil, fmt.Errorf("data file %s: explicit template %q not found", df.SourcePath, df.ExplicitTemplate)
			}
			if !fieldsCompatible(t, df, types) {
				return nil, fmt.Errorf("data file %s: explicit template %q requires fields not present in type %q", df.SourcePath, df.ExplicitTemplate, df.TypeName)
			}
			jobs = append(jobs, &Job{Data: df, Template: t})
			continue
		}

		found := false
		for _, orderName := range templateOrder {
			t, ok := tmplByName[orderName]
			if !ok {
				continue
			}
			if fieldsCompatible(t, df, types) {
				jobs = append(jobs, &Job{Data: df, Template: t})
				found = true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("data file %s: no compatible template found for type %q", df.SourcePath, df.TypeName)
		}
	}

	return jobs, nil
}

func fieldsCompatible(t *template.Template, df *data.DataFile, types map[string]*schema.TypeDefinition) bool {
	td, ok := types[df.TypeName]
	if !ok {
		return false
	}

	typeFieldSet := make(map[string]bool)
	for _, f := range td.Fields {
		typeFieldSet[f.Name] = true
	}

	for _, f := range t.RequiredFields {
		if !typeFieldSet[f] {
			return false
		}
	}

	return true
}
