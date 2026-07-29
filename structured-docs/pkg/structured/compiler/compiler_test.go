package compiler_test

import (
	"strings"
	"testing"

	"github.com/romayengineer/structured-docs/pkg/structured/compiler"
	"github.com/romayengineer/structured-docs/pkg/structured/config"
	"github.com/romayengineer/structured-docs/pkg/structured/fsys"
)

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
