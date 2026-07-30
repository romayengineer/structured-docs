//go:build integration

package compiler_test

import (
	"strings"
	"testing"

	"github.com/romayengineer/structured-docs/pkg/structured/compiler"
	"github.com/romayengineer/structured-docs/pkg/structured/config"
	"github.com/romayengineer/structured-docs/pkg/structured/fsys"
)

func boolPtr(b bool) *bool { return &b }

func writeProject(mem *fsys.MemFS, dataName, body string) {
	mem.WriteFile("schema/post.yml", []byte(`
fields:
  - name: title
    type: string
  - name: body
    type: string
`), 0644)
	mem.WriteFile("data/"+dataName, []byte(`
type: post
title: Test
body: `+body+`
`), 0644)
	mem.WriteFile("templates/post.template.md", []byte(`{{ .body }}`), 0644)
}

// --- File name validators ---

func TestIntegration_FileNameKebabPass(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "go-post.yml", "content")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			FileName: &config.FileNameRule{Style: "kebab"},
		},
	}
	_, err := compiler.Compile(mem, cfg)
	if err != nil {
		t.Fatalf("kebab pass expected no error, got: %v", err)
	}
}

func TestIntegration_FileNameKebabFail(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "GoPost.yml", "content")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			FileName: &config.FileNameRule{Style: "kebab"},
		},
	}
	_, err := compiler.Compile(mem, cfg)
	if err == nil {
		t.Fatal("kebab fail expected error, got nil")
	}
	if !strings.Contains(err.Error(), "does not match kebab style") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestIntegration_FileNameSnakePass(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "go_post.yml", "content")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			FileName: &config.FileNameRule{Style: "snake"},
		},
	}
	_, err := compiler.Compile(mem, cfg)
	if err != nil {
		t.Fatalf("snake pass expected no error, got: %v", err)
	}
}

func TestIntegration_FileNameSnakeFail(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "go-post.yml", "content")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			FileName: &config.FileNameRule{Style: "snake"},
		},
	}
	_, err := compiler.Compile(mem, cfg)
	if err == nil {
		t.Fatal("snake fail expected error, got nil")
	}
}

func TestIntegration_FileNameCamelPass(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "goPost.yml", "content")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			FileName: &config.FileNameRule{Style: "camel"},
		},
	}
	_, err := compiler.Compile(mem, cfg)
	if err != nil {
		t.Fatalf("camel pass expected no error, got: %v", err)
	}
}

func TestIntegration_FileNameCamelFail(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "go-post.yml", "content")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			FileName: &config.FileNameRule{Style: "camel"},
		},
	}
	_, err := compiler.Compile(mem, cfg)
	if err == nil {
		t.Fatal("camel fail expected error, got nil")
	}
}

func TestIntegration_FileNameLowercasePass(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "gopost.yml", "content")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			FileName: &config.FileNameRule{Style: "lowercase"},
		},
	}
	_, err := compiler.Compile(mem, cfg)
	if err != nil {
		t.Fatalf("lowercase pass expected no error, got: %v", err)
	}
}

func TestIntegration_FileNameLowercaseFail(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "go-post.yml", "content")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			FileName: &config.FileNameRule{Style: "lowercase"},
		},
	}
	_, err := compiler.Compile(mem, cfg)
	if err == nil {
		t.Fatal("lowercase fail expected error, got nil")
	}
}

func TestIntegration_FileNameCustomPattern(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "Mypost.yml", "content")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			FileName: &config.FileNameRule{Pattern: `^[A-Z][a-z]+$`},
		},
	}
	_, err := compiler.Compile(mem, cfg)
	if err != nil {
		t.Fatalf("custom pattern pass expected no error, got: %v", err)
	}
}

func TestIntegration_FileNameCustomPatternFail(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "myPost.yml", "content")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			FileName: &config.FileNameRule{Pattern: `^[A-Z][a-z]+$`},
		},
	}
	_, err := compiler.Compile(mem, cfg)
	if err == nil {
		t.Fatal("custom pattern fail expected error, got nil")
	}
}

func TestIntegration_FileNameNotConfigured(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "ANY-name.yml", "content")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
	}
	_, err := compiler.Compile(mem, cfg)
	if err != nil {
		t.Fatalf("no file_name config expected no error, got: %v", err)
	}
}

// --- Header spacing validators ---

func TestIntegration_HeaderSpacingPass(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "post.yml", "|\n  ## A\n  content\n\n  ## B")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			LinesBetweenHeaders: &config.HeaderSpacing{H2: 1},
		},
	}
	_, err := compiler.Compile(mem, cfg)
	if err != nil {
		t.Fatalf("header-spacing pass expected no error, got: %v", err)
	}
}

func TestIntegration_HeaderSpacingFail(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "post.yml", "|\n  ## A\n  content\n  ## B")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			LinesBetweenHeaders: &config.HeaderSpacing{H2: 1},
		},
	}
	_, err := compiler.Compile(mem, cfg)
	if err == nil {
		t.Fatal("header-spacing fail expected error, got nil")
	}
	if !strings.Contains(err.Error(), "expected 1 blank lines") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestIntegration_HeaderSpacingNotConfigured(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "post.yml", "|\n  ## A\n  content\n  ## B")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
	}
	_, err := compiler.Compile(mem, cfg)
	if err != nil {
		t.Fatalf("no header-spacing config expected no error, got: %v", err)
	}
}

// --- Validation blocks file write ---

func TestIntegration_FailDoesNotWriteOutput(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "BadName.yml", "content")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			FileName: &config.FileNameRule{Style: "kebab"},
		},
	}
	_, err := compiler.Compile(mem, cfg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	_, err = mem.ReadFile("output/BadName.md")
	if err == nil {
		t.Fatal("output file exists despite validation failure — write happened before validation")
	}
}

func TestIntegration_PassWritesOutput(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "good-name.yml", "content")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			FileName: &config.FileNameRule{Style: "kebab"},
		},
	}
	_, err := compiler.Compile(mem, cfg)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	b, err := mem.ReadFile("output/good-name.md")
	if err != nil {
		t.Fatal("output file not found — expected it to be written after validation passed")
	}
	if !strings.Contains(string(b), "content") {
		t.Errorf("output missing expected content, got: %s", string(b))
	}
}

// --- Multiple validators ---

func TestIntegration_BothValidatorsPass(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "good.yml", "|\n  ## A\n  content\n\n  ## B")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			LinesBetweenHeaders: &config.HeaderSpacing{H2: 1},
			FileName:            &config.FileNameRule{Style: "kebab"},
		},
	}
	_, err := compiler.Compile(mem, cfg)
	if err != nil {
		t.Fatalf("both validators pass expected no error, got: %v", err)
	}
}

func TestIntegration_BothValidators_FirstFails(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "BAD.yml", "|\n  ## A\n  content\n\n  ## B")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			LinesBetweenHeaders: &config.HeaderSpacing{H2: 1},
			FileName:            &config.FileNameRule{Style: "kebab"},
		},
	}
	_, err := compiler.Compile(mem, cfg)
	if err == nil {
		t.Fatal("expected error from file-name validator, got nil")
	}
	if !strings.Contains(err.Error(), "does not match") {
		t.Errorf("expected file-name error, got: %v", err)
	}
}

func TestIntegration_BothValidators_SecondFails(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "good.yml", "|\n  ## A\n  content\n  ## B")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			LinesBetweenHeaders: &config.HeaderSpacing{H2: 1},
			FileName:            &config.FileNameRule{Style: "kebab"},
		},
	}
	_, err := compiler.Compile(mem, cfg)
	if err == nil {
		t.Fatal("expected error from header-spacing validator, got nil")
	}
	if !strings.Contains(err.Error(), "expected 1 blank lines") {
		t.Errorf("expected header-spacing error, got: %v", err)
	}
}

// --- Combined: file name passes but spacing fails (second validator catches) ---

func TestIntegration_NoValidateConfigSkipsAll(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "BAD.yml", "|\n  ## A\n  content\n  ## B")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
	}
	_, err := compiler.Compile(mem, cfg)
	if err != nil {
		t.Fatalf("no validate config should skip all validators, got: %v", err)
	}
}

// --- Leading digit ---

func TestIntegration_LeadingDigitKebabPass(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "go-post.yml", "content")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			FileName: &config.FileNameRule{
				Style:             "kebab",
				AllowLeadingDigit: boolPtr(false),
			},
		},
	}
	_, err := compiler.Compile(mem, cfg)
	if err != nil {
		t.Fatalf("kebab+no-leading-digit pass expected no error, got: %v", err)
	}
}

func TestIntegration_LeadingDigitKebabFail(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "123-post.yml", "content")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			FileName: &config.FileNameRule{
				Style:             "kebab",
				AllowLeadingDigit: boolPtr(false),
			},
		},
	}
	_, err := compiler.Compile(mem, cfg)
	if err == nil {
		t.Fatal("expected error for leading digit, got nil")
	}
	if !strings.Contains(err.Error(), "must not start with a digit") {
		t.Errorf("expected leading digit error, got: %v", err)
	}
}

func TestIntegration_LeadingDigitDefaultAllows(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "123-post.yml", "content")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			FileName: &config.FileNameRule{
				Style: "kebab",
			},
		},
	}
	_, err := compiler.Compile(mem, cfg)
	if err != nil {
		t.Fatalf("default (no allow_leading_digit) should allow digits, got: %v", err)
	}
}

func TestIntegration_LeadingDigitWithSnake(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "123_post.yml", "content")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			FileName: &config.FileNameRule{
				Style:             "snake",
				AllowLeadingDigit: boolPtr(false),
			},
		},
	}
	_, err := compiler.Compile(mem, cfg)
	if err == nil {
		t.Fatal("expected error for leading digit with snake style, got nil")
	}
}

func TestIntegration_LeadingDigitBlocksWrite(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "1bad.yml", "content")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
		Validate: &config.Validate{
			FileName: &config.FileNameRule{
				AllowLeadingDigit: boolPtr(false),
			},
		},
	}
	_, err := compiler.Compile(mem, cfg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	_, err = mem.ReadFile("output/1bad.md")
	if err == nil {
		t.Fatal("output file exists despite validation failure")
	}
}

func TestIntegration_ValidateNilConfigSkips(t *testing.T) {
	mem := fsys.NewMemFS()
	writeProject(mem, "BAD.yml", "|\n  ## A\n  content\n  ## B")
	cfg := &config.Config{
		SchemaDir: "schema", DataDir: "data", TemplateDir: "templates",
		OutputDir: "output", TemplateOrder: []string{"post.template.md"},
		Validate:  &config.Validate{},
	}
	_, err := compiler.Compile(mem, cfg)
	if err != nil {
		t.Fatalf("empty validate config should skip all validators, got: %v", err)
	}
}
