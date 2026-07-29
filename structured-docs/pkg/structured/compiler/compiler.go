package compiler

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/romayengineer/structured-docs/pkg/structured/config"
	"github.com/romayengineer/structured-docs/pkg/structured/data"
	"github.com/romayengineer/structured-docs/pkg/structured/fsys"
	"github.com/romayengineer/structured-docs/pkg/structured/renderer"
	"github.com/romayengineer/structured-docs/pkg/structured/resolver"
	"github.com/romayengineer/structured-docs/pkg/structured/schema"
	"github.com/romayengineer/structured-docs/pkg/structured/template"
)

type Result struct {
	SourcePath string
	OutputPath string
	Format     string
}

func Compile(fsys fsys.FS, cfg *config.Config) ([]Result, error) {
	types, err := schema.LoadAll(fsys, cfg.SchemaDir)
	if err != nil {
		return nil, fmt.Errorf("loading schemas: %w", err)
	}

	templates, err := template.LoadAll(fsys, cfg.TemplateDir)
	if err != nil {
		return nil, fmt.Errorf("loading templates: %w", err)
	}

	dataFiles, err := data.LoadAll(fsys, cfg.DataDir, types)
	if err != nil {
		return nil, fmt.Errorf("loading data: %w", err)
	}

	jobs, err := resolver.ResolveAll(dataFiles, templates, cfg.TemplateOrder, types)
	if err != nil {
		return nil, fmt.Errorf("resolving templates: %w", err)
	}

	var results []Result

	for _, job := range jobs {
		output, err := renderer.Render(job)
		if err != nil {
			return nil, fmt.Errorf("rendering %s: %w", job.Data.SourcePath, err)
		}

		relDir := filepath.Dir(job.Data.SourcePath)
		base := strings.TrimSuffix(filepath.Base(job.Data.SourcePath), ".yml")
		base = strings.TrimSuffix(base, ".yaml")

		outName := base + job.Template.OutputExt
		outPath := filepath.Join(cfg.OutputDir, relDir, outName)

		if err := fsys.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return nil, fmt.Errorf("creating output directory for %s: %w", outPath, err)
		}

		if err := fsys.WriteFile(outPath, []byte(output), 0644); err != nil {
			return nil, fmt.Errorf("writing output %s: %w", outPath, err)
		}

		results = append(results, Result{
			SourcePath: job.Data.SourcePath,
			OutputPath: outPath,
			Format:     strings.TrimPrefix(job.Template.OutputExt, "."),
		})
	}

	return results, nil
}

func CleanOutput(fsys fsys.FS, cfg *config.Config) error {
	return fsys.RemoveAll(cfg.OutputDir)
}
