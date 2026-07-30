//go:build integration

package compiler_test

import (
	"strings"
	"testing"

	"github.com/romayengineer/structured-docs/pkg/structured/compiler"
	"github.com/romayengineer/structured-docs/pkg/structured/config"
	"github.com/romayengineer/structured-docs/pkg/structured/fsys"
)

func TestIntegration_FileNameValidationPass(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("schema/post.yml", []byte(`
fields:
  - name: title
    type: string
`), 0644)
	mem.WriteFile("data/go-post.yml", []byte(`
type: post
title: Hello
`), 0644)
	mem.WriteFile("templates/post.template.md", []byte(`{{ .title }}`), 0644)

	cfg := &config.Config{
		SchemaDir:     "schema",
		DataDir:       "data",
		TemplateDir:   "templates",
		OutputDir:     "output",
		TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			FileName: &config.FileNameRule{Style: "kebab"},
		},
	}

	_, err := compiler.Compile(mem, cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestIntegration_FileNameValidationFail(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("schema/post.yml", []byte(`
fields:
  - name: title
    type: string
`), 0644)
	mem.WriteFile("data/TestPost.yml", []byte(`
type: post
title: Hello
`), 0644)
	mem.WriteFile("templates/post.template.md", []byte(`{{ .title }}`), 0644)

	cfg := &config.Config{
		SchemaDir:     "schema",
		DataDir:       "data",
		TemplateDir:   "templates",
		OutputDir:     "output",
		TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			FileName: &config.FileNameRule{Style: "kebab"},
		},
	}

	_, err := compiler.Compile(mem, cfg)
	if err == nil {
		t.Fatal("expected validation error for non-kebab file name, got nil")
	}
	if !strings.Contains(err.Error(), "does not match kebab style") {
		t.Errorf("expected kebab style error, got: %v", err)
	}
}
