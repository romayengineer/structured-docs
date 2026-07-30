package config_test

import (
	"testing"

	"github.com/romayengineer/structured-docs/pkg/structured/config"
	"github.com/romayengineer/structured-docs/pkg/structured/fsys"
)

func TestDefault(t *testing.T) {
	cfg := config.Default()
	if cfg.SchemaDir != "schema" {
		t.Errorf("expected schema, got %q", cfg.SchemaDir)
	}
	if cfg.DataDir != "data" {
		t.Errorf("expected data, got %q", cfg.DataDir)
	}
	if cfg.TemplateDir != "templates" {
		t.Errorf("expected templates, got %q", cfg.TemplateDir)
	}
	if cfg.OutputDir != "output" {
		t.Errorf("expected output, got %q", cfg.OutputDir)
	}
	if cfg.TemplateOrder != nil {
		t.Errorf("expected nil template order, got %v", cfg.TemplateOrder)
	}
}

func TestLoad_Success(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("structured.yml", []byte(`
schema_dir: myschema
data_dir: mydata
template_dir: mytemplates
output_dir: myoutput
template_order:
  - post.md
`), 0644)

	cfg, err := config.Load(mem, "structured.yml")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SchemaDir != "myschema" {
		t.Errorf("expected myschema, got %q", cfg.SchemaDir)
	}
	if cfg.DataDir != "mydata" {
		t.Errorf("expected mydata, got %q", cfg.DataDir)
	}
	if cfg.TemplateDir != "mytemplates" {
		t.Errorf("expected mytemplates, got %q", cfg.TemplateDir)
	}
	if cfg.OutputDir != "myoutput" {
		t.Errorf("expected myoutput, got %q", cfg.OutputDir)
	}
	if len(cfg.TemplateOrder) != 1 || cfg.TemplateOrder[0] != "post.md" {
		t.Errorf("expected [post.md], got %v", cfg.TemplateOrder)
	}
}

func TestLoad_PartialOverride(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("config.yml", []byte(`
template_order:
  - page.md
`), 0644)

	cfg, err := config.Load(mem, "config.yml")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SchemaDir != "schema" {
		t.Errorf("expected default schema, got %q", cfg.SchemaDir)
	}
	if len(cfg.TemplateOrder) != 1 || cfg.TemplateOrder[0] != "page.md" {
		t.Errorf("expected [page.md], got %v", cfg.TemplateOrder)
	}
}

func TestLoad_ReadError(t *testing.T) {
	mem := fsys.NewMemFS()
	_, err := config.Load(mem, "nonexistent.yml")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoad_ParseError(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("bad.yml", []byte(`{{{invalid yaml`), 0644)

	_, err := config.Load(mem, "bad.yml")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoad_MissingTemplateOrder(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("no-order.yml", []byte(`schema_dir: s`), 0644)

	_, err := config.Load(mem, "no-order.yml")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
