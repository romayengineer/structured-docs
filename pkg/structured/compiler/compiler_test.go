package compiler_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/romayengineer/structured-docs/pkg/structured/compiler"
	"github.com/romayengineer/structured-docs/pkg/structured/config"
	"github.com/romayengineer/structured-docs/pkg/structured/data"
	"github.com/romayengineer/structured-docs/pkg/structured/fsys"
	"github.com/romayengineer/structured-docs/pkg/structured/resolver"
	"github.com/romayengineer/structured-docs/pkg/structured/schema"
	"github.com/romayengineer/structured-docs/pkg/structured/template"
)

type failMkdirAllFS struct {
	fsys.FS
}

func (f *failMkdirAllFS) MkdirAll(string, os.FileMode) error {
	return fmt.Errorf("mock mkdir error")
}

type failWriteFileFS struct {
	fsys.FS
}

func (f *failWriteFileFS) WriteFile(string, []byte, os.FileMode) error {
	return fmt.Errorf("mock write error")
}

func writeExampleProject(mem *fsys.MemFS) {
	mem.WriteFile("schema/post.yml", []byte(`
description: A blog post
fields:
  - name: title
    type: string
    required: true
  - name: date
    type: string
  - name: body
    type: string
    required: true
  - name: tags
    type: "[]string"
`), 0644)

	mem.WriteFile("data/blog/hello.yml", []byte(`
type: post
title: Hello World
date: "2026-07-29"
body: This is my first post!
tags: [hello, world]
`), 0644)

	mem.WriteFile("templates/post.template.md", []byte(`
# {{ .title }}
{{ .body }}
{{ range .tags }} {{ . }}{{ end }}
`), 0644)
}

func TestEndToEnd(t *testing.T) {
	mem := fsys.NewMemFS()
	writeExampleProject(mem)

	cfg := &config.Config{
		SchemaDir:     "schema",
		DataDir:       "data",
		TemplateDir:   "templates",
		OutputDir:     "output",
		TemplateOrder: []string{"post.template.md"},
	}

	results, err := compiler.Compile(mem, cfg)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.SourcePath != "blog/hello.yml" {
		t.Errorf("expected source path blog/hello.yml, got %s", r.SourcePath)
	}
	if r.OutputPath != "output/blog/hello.md" {
		t.Errorf("expected output path output/blog/hello.md, got %s", r.OutputPath)
	}
	if r.Format != "md" {
		t.Errorf("expected format md, got %s", r.Format)
	}

	b, err := mem.ReadFile("output/blog/hello.md")
	if err != nil {
		t.Fatal(err)
	}

	content := string(b)
	if !strings.Contains(content, "Hello World") {
		t.Errorf("output should contain title")
	}
	if !strings.Contains(content, "This is my first post!") {
		t.Errorf("output should contain body")
	}
}

func TestMissingRequiredField(t *testing.T) {
	mem := fsys.NewMemFS()

	mem.WriteFile("schema/post.yml", []byte(`
description: A blog post
fields:
  - name: title
    type: string
    required: true
`), 0644)

	mem.WriteFile("data/post.yml", []byte(`
type: post
`), 0644)

	mem.WriteFile("templates/post.template.md", []byte(`{{ .title }}`), 0644)

	cfg := &config.Config{
		SchemaDir:     "schema",
		DataDir:       "data",
		TemplateDir:   "templates",
		OutputDir:     "output",
		TemplateOrder: []string{"post.template.md"},
	}

	_, err := compiler.Compile(mem, cfg)
	if err == nil {
		t.Fatal("expected error for missing required field, got nil")
	}
	if !strings.Contains(err.Error(), "title") {
		t.Errorf("error should mention missing field 'title', got: %v", err)
	}
}

func TestUnknownType(t *testing.T) {
	mem := fsys.NewMemFS()

	mem.WriteFile("schema/post.yml", []byte(`
fields:
  - name: title
    type: string
`), 0644)

	mem.WriteFile("data/post.yml", []byte(`
type: unknown
title: hi
`), 0644)

	mem.WriteFile("templates/post.template.md", []byte(`{{ .title }}`), 0644)

	cfg := &config.Config{
		SchemaDir:     "schema",
		DataDir:       "data",
		TemplateDir:   "templates",
		OutputDir:     "output",
		TemplateOrder: []string{"post.template.md"},
	}

	_, err := compiler.Compile(mem, cfg)
	if err == nil {
		t.Fatal("expected error for unknown type, got nil")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("error should mention unknown type, got: %v", err)
	}
}

func TestNoCompatibleTemplate(t *testing.T) {
	mem := fsys.NewMemFS()

	mem.WriteFile("schema/post.yml", []byte(`
fields:
  - name: title
    type: string
`), 0644)

	mem.WriteFile("data/post.yml", []byte(`
type: post
title: hi
`), 0644)

	mem.WriteFile("templates/longform.template.md", []byte(`{{ .title }} {{ .body }}`), 0644)

	cfg := &config.Config{
		SchemaDir:     "schema",
		DataDir:       "data",
		TemplateDir:   "templates",
		OutputDir:     "output",
		TemplateOrder: []string{"longform.template.md"},
	}

	_, err := compiler.Compile(mem, cfg)
	if err == nil {
		t.Fatal("expected error for no compatible template, got nil")
	}
}

func TestCleanOutput(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.MkdirAll("output/blog", 0755)
	mem.WriteFile("output/blog/hello.md", []byte("content"), 0644)

	cfg := &config.Config{OutputDir: "output"}

	if err := compiler.CleanOutput(mem, cfg); err != nil {
		t.Fatal(err)
	}

	_, err := mem.ReadFile("output/blog/hello.md")
	if err == nil {
		t.Error("expected file to be removed")
	}
}

func TestFieldTypeCoercion(t *testing.T) {
	mem := fsys.NewMemFS()

	mem.WriteFile("schema/page.yml", []byte(`
fields:
  - name: count
    type: int
  - name: ratio
    type: float
  - name: active
    type: bool
`), 0644)

	mem.WriteFile("data/page.yml", []byte(`
type: page
count: 42
ratio: 3.14
active: true
`), 0644)

	mem.WriteFile("templates/page.template.md", []byte(
		`count={{ .count }} ratio={{ .ratio }} active={{ .active }}`), 0644)

	cfg := &config.Config{
		SchemaDir:     "schema",
		DataDir:       "data",
		TemplateDir:   "templates",
		OutputDir:     "output",
		TemplateOrder: []string{"page.template.md"},
	}

	results, err := compiler.Compile(mem, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	b, _ := mem.ReadFile("output/page.md")
	content := string(b)
	if !strings.Contains(content, "count=42") {
		t.Errorf("expected count=42, got: %s", content)
	}
	if !strings.Contains(content, "ratio=3.14") {
		t.Errorf("expected ratio=3.14, got: %s", content)
	}
	if !strings.Contains(content, "active=true") {
		t.Errorf("expected active=true, got: %s", content)
	}
}

func TestExplicitTemplate(t *testing.T) {
	mem := fsys.NewMemFS()

	mem.WriteFile("schema/post.yml", []byte(`
fields:
  - name: title
    type: string
`), 0644)

	mem.WriteFile("data/post.yml", []byte(`
type: post
template: custom.template.md
title: Hello
`), 0644)

	mem.WriteFile("templates/default.template.md", []byte(`default: {{ .title }}`), 0644)
	mem.WriteFile("templates/custom.template.md", []byte(`custom: {{ .title }}`), 0644)

	cfg := &config.Config{
		SchemaDir:     "schema",
		DataDir:       "data",
		TemplateDir:   "templates",
		OutputDir:     "output",
		TemplateOrder: []string{"default.template.md"},
	}

	_, err := compiler.Compile(mem, cfg)
	if err != nil {
		t.Fatal(err)
	}

	b, _ := mem.ReadFile("output/post.md")
	if !strings.Contains(string(b), "custom:") {
		t.Errorf("expected custom template, got: %s", string(b))
	}
}

func TestOverrideRenderer(t *testing.T) {
	c := &compiler.Compiler{
		FS: fsys.NewMemFS(),
		Schema: compiler.SchemaLoaderFunc(func(_ fsys.FS, _ string) (map[string]*schema.TypeDefinition, error) {
			return map[string]*schema.TypeDefinition{
				"post": {Name: "post", Fields: []schema.FieldDefinition{
					{Name: "title", Type: "string", Required: schema.BoolPtr(true)},
				}},
			}, nil
		}),
		Template: compiler.TemplateLoaderFunc(func(_ fsys.FS, _ string) ([]*template.Template, error) {
			return []*template.Template{
				{FileName: "post.template.md", Content: "{{ .title }}", RequiredFields: []string{"title"}, OutputExt: ".md"},
			}, nil
		}),
		Data: compiler.DataLoaderFunc(func(_ fsys.FS, _ string, _ map[string]*schema.TypeDefinition) ([]*data.DataFile, error) {
			return []*data.DataFile{
				{SourcePath: "test.yml", TypeName: "post", Fields: map[string]interface{}{"title": "Hello"}},
			}, nil
		}),
		Resolve: compiler.ResolverFunc(func(df []*data.DataFile, _ []*template.Template, order []string, _ map[string]*schema.TypeDefinition) ([]*resolver.Job, error) {
			var jobs []*resolver.Job
			for _, d := range df {
				jobs = append(jobs, &resolver.Job{Data: d, Template: &template.Template{FileName: order[0], OutputExt: ".md"}})
			}
			return jobs, nil
		}),
		Render: compiler.RendererFunc(func(job *resolver.Job) (string, error) {
			return "mocked-render-output", nil
		}),
	}

	cfg := &config.Config{
		OutputDir:     "output",
		TemplateOrder: []string{"post.template.md"},
	}

	results, err := c.Compile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.OutputPath != "output/test.md" {
		t.Errorf("expected output/test.md, got %s", r.OutputPath)
	}

	b, _ := c.FS.ReadFile("output/test.md")
	if string(b) != "mocked-render-output" {
		t.Errorf("expected 'mocked-render-output', got %q", string(b))
	}
}

func TestOverrideSchemaLoader(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("data/post.yml", []byte("type: custom\ntitle: Hello\n"), 0644)
	mem.WriteFile("templates/custom.template.md", []byte("{{ .title }}"), 0644)

	c := compiler.New(mem)
	c.Schema = compiler.SchemaLoaderFunc(func(fs fsys.FS, dir string) (map[string]*schema.TypeDefinition, error) {
		return map[string]*schema.TypeDefinition{
	"custom": {
			Name: "custom",
			Fields: []schema.FieldDefinition{
				{Name: "title", Type: "string", Required: schema.BoolPtr(true)},
			},
			},
		}, nil
	})

	cfg := &config.Config{
		SchemaDir:     "schema",
		DataDir:       "data",
		TemplateDir:   "templates",
		OutputDir:     "output",
		TemplateOrder: []string{"custom.template.md"},
	}

	results, err := c.Compile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	b, _ := mem.ReadFile("output/post.md")
	if string(b) != "Hello" {
		t.Errorf("expected 'Hello', got %q", string(b))
	}
}

func TestOverrideResolverFails(t *testing.T) {
	c := &compiler.Compiler{
		FS: fsys.NewMemFS(),
		Schema: compiler.SchemaLoaderFunc(func(_ fsys.FS, _ string) (map[string]*schema.TypeDefinition, error) {
			return map[string]*schema.TypeDefinition{"post": {Name: "post"}}, nil
		}),
		Template: compiler.TemplateLoaderFunc(func(_ fsys.FS, _ string) ([]*template.Template, error) {
			return nil, nil
		}),
		Data: compiler.DataLoaderFunc(func(_ fsys.FS, _ string, _ map[string]*schema.TypeDefinition) ([]*data.DataFile, error) {
			return nil, nil
		}),
		Resolve: compiler.ResolverFunc(func(
			_ []*data.DataFile,
			_ []*template.Template,
			_ []string,
			_ map[string]*schema.TypeDefinition,
		) ([]*resolver.Job, error) {
			return nil, fmt.Errorf("mock resolver failure")
		}),
	}

	_, err := c.Compile(&config.Config{TemplateOrder: []string{"any.md"}})
	if err == nil {
		t.Fatal("expected error from mock resolver, got nil")
	}
	if !strings.Contains(err.Error(), "mock resolver failure") {
		t.Errorf("expected 'mock resolver failure', got: %v", err)
	}
}

func TestCompile_SchemaError(t *testing.T) {
	mem := fsys.NewMemFS()
	writeExampleProject(mem)
	c := compiler.New(mem)
	c.Schema = compiler.SchemaLoaderFunc(func(_ fsys.FS, _ string) (map[string]*schema.TypeDefinition, error) {
		return nil, fmt.Errorf("mock schema error")
	})

	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
	}
	_, err := c.Compile(cfg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "mock schema error") {
		t.Errorf("expected mock schema error, got: %v", err)
	}
}

func TestCompile_TemplateError(t *testing.T) {
	mem := fsys.NewMemFS()
	writeExampleProject(mem)
	c := compiler.New(mem)
	c.Template = compiler.TemplateLoaderFunc(func(_ fsys.FS, _ string) ([]*template.Template, error) {
		return nil, fmt.Errorf("mock template error")
	})

	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
	}
	_, err := c.Compile(cfg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "mock template error") {
		t.Errorf("expected mock template error, got: %v", err)
	}
}

func TestCompile_DataError(t *testing.T) {
	mem := fsys.NewMemFS()
	writeExampleProject(mem)
	c := compiler.New(mem)
	c.Data = compiler.DataLoaderFunc(func(_ fsys.FS, _ string, _ map[string]*schema.TypeDefinition) ([]*data.DataFile, error) {
		return nil, fmt.Errorf("mock data error")
	})

	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
	}
	_, err := c.Compile(cfg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "mock data error") {
		t.Errorf("expected mock data error, got: %v", err)
	}
}

func TestCompile_RenderError(t *testing.T) {
	mem := fsys.NewMemFS()
	writeExampleProject(mem)
	c := compiler.New(mem)
	c.Render = compiler.RendererFunc(func(_ *resolver.Job) (string, error) {
		return "", fmt.Errorf("mock render error")
	})

	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
	}
	_, err := c.Compile(cfg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "mock render error") {
		t.Errorf("expected mock render error, got: %v", err)
	}
}

func TestCompile_MkdirAllError(t *testing.T) {
	mem := fsys.NewMemFS()
	writeExampleProject(mem)
	efs := &failMkdirAllFS{FS: mem}

	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
	}
	_, err := compiler.Compile(efs, cfg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "mock mkdir error") {
		t.Errorf("expected mock mkdir error, got: %v", err)
	}
}

func TestCompile_WriteFileError(t *testing.T) {
	mem := fsys.NewMemFS()
	writeExampleProject(mem)

	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
	}

	efs := &failWriteFileFS{FS: mem}
	_, err := compiler.Compile(efs, cfg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "mock write error") {
		t.Errorf("expected mock write error, got: %v", err)
	}
}
