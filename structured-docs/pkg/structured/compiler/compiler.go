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

type SchemaLoader   func(fsys.FS, string) (map[string]*schema.TypeDefinition, error)
type TemplateLoader func(fsys.FS, string) ([]*template.Template, error)
type DataLoader     func(fsys.FS, string, map[string]*schema.TypeDefinition) ([]*data.DataFile, error)
type Resolver      func([]*data.DataFile, []*template.Template, []string, map[string]*schema.TypeDefinition) ([]*resolver.Job, error)
type Renderer      func(*resolver.Job) (string, error)

type Compiler struct {
	FS       fsys.FS
	Schema   SchemaLoader
	Template TemplateLoader
	Data     DataLoader
	Resolve  Resolver
	Render   Renderer
}

func New(fs fsys.FS) *Compiler {
	return &Compiler{
		FS:       fs,
		Schema:   schema.LoadAll,
		Template: template.LoadAll,
		Data:     data.LoadAll,
		Resolve:  resolver.ResolveAll,
		Render:   renderer.Render,
	}
}

func (c *Compiler) Compile(cfg *config.Config) ([]Result, error) {
	types, err := c.Schema(c.FS, cfg.SchemaDir)
	if err != nil {
		return nil, fmt.Errorf("loading schemas: %w", err)
	}

	templates, err := c.Template(c.FS, cfg.TemplateDir)
	if err != nil {
		return nil, fmt.Errorf("loading templates: %w", err)
	}

	dataFiles, err := c.Data(c.FS, cfg.DataDir, types)
	if err != nil {
		return nil, fmt.Errorf("loading data: %w", err)
	}

	jobs, err := c.Resolve(dataFiles, templates, cfg.TemplateOrder, types)
	if err != nil {
		return nil, fmt.Errorf("resolving templates: %w", err)
	}

	var results []Result

	for _, job := range jobs {
		output, err := c.Render(job)
		if err != nil {
			return nil, fmt.Errorf("rendering %s: %w", job.Data.SourcePath, err)
		}

		relDir := filepath.Dir(job.Data.SourcePath)
		base := strings.TrimSuffix(filepath.Base(job.Data.SourcePath), ".yml")
		base = strings.TrimSuffix(base, ".yaml")

		outName := base + job.Template.OutputExt
		outPath := filepath.Join(cfg.OutputDir, relDir, outName)

		if err := c.FS.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return nil, fmt.Errorf("creating output directory for %s: %w", outPath, err)
		}

		if err := c.FS.WriteFile(outPath, []byte(output), 0644); err != nil {
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

func (c *Compiler) CleanOutput(cfg *config.Config) error {
	return c.FS.RemoveAll(cfg.OutputDir)
}

type Result struct {
	SourcePath string
	OutputPath string
	Format     string
}

func Compile(fs fsys.FS, cfg *config.Config) ([]Result, error) {
	return New(fs).Compile(cfg)
}

func CleanOutput(fs fsys.FS, cfg *config.Config) error {
	return New(fs).CleanOutput(cfg)
}
