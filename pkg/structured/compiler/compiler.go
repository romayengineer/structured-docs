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
	"github.com/romayengineer/structured-docs/pkg/structured/validator"
)

type SchemaLoader interface {
	LoadSchema(fsys.FS, string) (map[string]*schema.TypeDefinition, error)
}

type SchemaLoaderFunc func(fsys.FS, string) (map[string]*schema.TypeDefinition, error)

func (f SchemaLoaderFunc) LoadSchema(fs fsys.FS, dir string) (map[string]*schema.TypeDefinition, error) {
	return f(fs, dir)
}

type TemplateLoader interface {
	LoadTemplates(fsys.FS, string) ([]*template.Template, error)
}

type TemplateLoaderFunc func(fsys.FS, string) ([]*template.Template, error)

func (f TemplateLoaderFunc) LoadTemplates(fs fsys.FS, dir string) ([]*template.Template, error) {
	return f(fs, dir)
}

type DataLoader interface {
	LoadData(fsys.FS, string, map[string]*schema.TypeDefinition) ([]*data.DataFile, error)
}

type DataLoaderFunc func(fsys.FS, string, map[string]*schema.TypeDefinition) ([]*data.DataFile, error)

func (f DataLoaderFunc) LoadData(fs fsys.FS, dir string, types map[string]*schema.TypeDefinition) ([]*data.DataFile, error) {
	return f(fs, dir, types)
}

type Resolver interface {
	Resolve([]*data.DataFile, []*template.Template, []string, map[string]*schema.TypeDefinition) ([]*resolver.Job, error)
}

type ResolverFunc func([]*data.DataFile, []*template.Template, []string, map[string]*schema.TypeDefinition) ([]*resolver.Job, error)

func (f ResolverFunc) Resolve(dataFiles []*data.DataFile, templates []*template.Template, templateOrder []string, types map[string]*schema.TypeDefinition) ([]*resolver.Job, error) {
	return f(dataFiles, templates, templateOrder, types)
}

type Renderer interface {
	Render(*resolver.Job) (string, error)
}

type RendererFunc func(*resolver.Job) (string, error)

func (f RendererFunc) Render(job *resolver.Job) (string, error) {
	return f(job)
}

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
		Schema:   SchemaLoaderFunc(schema.LoadAll),
		Template: TemplateLoaderFunc(template.LoadAll),
		Data:     DataLoaderFunc(data.LoadAll),
		Resolve:  ResolverFunc(resolver.ResolveAll),
		Render:   RendererFunc(renderer.Render),
	}
}

func (c *Compiler) Compile(cfg *config.Config) ([]Result, error) {
	types, err := c.Schema.LoadSchema(c.FS, cfg.SchemaDir)
	if err != nil {
		return nil, fmt.Errorf("loading schemas: %w", err)
	}

	templates, err := c.Template.LoadTemplates(c.FS, cfg.TemplateDir)
	if err != nil {
		return nil, fmt.Errorf("loading templates: %w", err)
	}

	dataFiles, err := c.Data.LoadData(c.FS, cfg.DataDir, types)
	if err != nil {
		return nil, fmt.Errorf("loading data: %w", err)
	}

	jobs, err := c.Resolve.Resolve(dataFiles, templates, cfg.TemplateOrder, types)
	if err != nil {
		return nil, fmt.Errorf("resolving templates: %w", err)
	}

	var results []Result

	for _, job := range jobs {
		output, err := c.Render.Render(job)
		if err != nil {
			return nil, fmt.Errorf("rendering %s: %w", job.Data.SourcePath, err)
		}

		relDir := filepath.Dir(job.Data.SourcePath)
		base := strings.TrimSuffix(filepath.Base(job.Data.SourcePath), ".yml")
		base = strings.TrimSuffix(base, ".yaml")

		outName := base + job.Template.OutputExt
		outPath := filepath.Join(cfg.OutputDir, relDir, outName)

		if cfg.Validate != nil {
			for _, v := range validator.DefaultValidators() {
				if errs := v.Validate(output, outPath, cfg.Validate); len(errs) > 0 {
					return nil, fmt.Errorf("validation: %w", errs[0])
				}
			}
		}

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
	outputAbs, err := filepath.Abs(cfg.OutputDir)
	if err != nil {
		return fmt.Errorf("resolving output dir: %w", err)
	}

	var sourceDirs []string
	for _, dir := range []string{cfg.SchemaDir, cfg.DataDir, cfg.TemplateDir} {
		if dir == "" {
			continue
		}
		abs, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("resolving %s: %w", dir, err)
		}
		sourceDirs = append(sourceDirs, abs)
	}

	needsSmartClean := false
	for _, src := range sourceDirs {
		if isInside(src, outputAbs) {
			needsSmartClean = true
			break
		}
	}

	if !needsSmartClean {
		return c.FS.RemoveAll(cfg.OutputDir)
	}

	entries, err := c.FS.ReadDir(cfg.OutputDir)
	if err != nil {
		return nil
	}

	for _, e := range entries {
		if e.Name() == ".git" || e.Name() == ".gitignore" {
			continue
		}

		protect := false
		for _, src := range sourceDirs {
			if isInside(src, filepath.Join(outputAbs, e.Name())) {
				protect = true
				break
			}
		}
		if protect {
			continue
		}

		if err := c.FS.RemoveAll(filepath.Join(cfg.OutputDir, e.Name())); err != nil {
			return err
		}
	}

	return nil
}

func isInside(child, parent string) bool {
	parent = filepath.Clean(parent)
	child = filepath.Clean(child)
	if parent == child {
		return true
	}
	prefix := parent + string(filepath.Separator)
	return strings.HasPrefix(child, prefix)
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
