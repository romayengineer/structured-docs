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

func TestLoad_WithValidate(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("validated.yml", []byte(`
schema_dir: schema
data_dir: data
template_dir: templates
output_dir: output
template_order:
  - post.template.md
validate:
  lines_between_headers:
    h2: 2
    h3: 1
    default: 1
`), 0644)

	cfg, err := config.Load(mem, "validated.yml")
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Validate == nil {
		t.Fatal("expected validate config, got nil")
	}
	if cfg.Validate.LinesBetweenHeaders == nil {
		t.Fatal("expected LinesBetweenHeaders config, got nil")
	}

	s := cfg.Validate.LinesBetweenHeaders
	if s.H2 != 2 {
		t.Errorf("expected H2=2, got %d", s.H2)
	}
	if s.H3 != 1 {
		t.Errorf("expected H3=1, got %d", s.H3)
	}
	if s.Default != 1 {
		t.Errorf("expected Default=1, got %d", s.Default)
	}
	if s.H1 != 0 {
		t.Errorf("expected H1=0 (unset), got %d", s.H1)
	}
}

func TestLoad_WithFileRule(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("validated.yml", []byte(`
template_order:
  - post.template.md
validate:
  file_name:
    style: kebab
`), 0644)

	cfg, err := config.Load(mem, "validated.yml")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Validate == nil {
		t.Fatal("expected validate config, got nil")
	}
	if cfg.Validate.FileName == nil {
		t.Fatal("expected FileName config, got nil")
	}
	if cfg.Validate.FileName.Style != "kebab" {
		t.Errorf("expected style=kebab, got %q", cfg.Validate.FileName.Style)
	}
}

func TestLoad_WithAllowLeadingDigit(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("validated.yml", []byte(`
template_order:
  - post.template.md
validate:
  file_name:
    style: kebab
    allow_leading_digit: false
`), 0644)

	cfg, err := config.Load(mem, "validated.yml")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Validate.FileName == nil {
		t.Fatal("expected FileName config, got nil")
	}
	if cfg.Validate.FileName.AllowLeadingDigit == nil {
		t.Fatal("expected AllowLeadingDigit to be set, got nil")
	}
	if *cfg.Validate.FileName.AllowLeadingDigit {
		t.Error("expected AllowLeadingDigit=false, got true")
	}
}

func TestLoad_AllowLeadingDigitDefaultIsNil(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("no-ald.yml", []byte(`
template_order:
  - post.template.md
validate:
  file_name:
    style: kebab
`), 0644)

	cfg, err := config.Load(mem, "no-ald.yml")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Validate.FileName.AllowLeadingDigit != nil {
		t.Error("expected AllowLeadingDigit=nil when not set, got non-nil")
	}
}

func TestLoad_GenerateReadmeDefault(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("config.yml", []byte(`
template_order:
  - post.template.md
`), 0644)

	cfg, err := config.Load(mem, "config.yml")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GenerateReadme != nil {
		t.Error("expected GenerateReadme=nil when not set, got non-nil")
	}
}

func TestLoad_GenerateReadmeExplicit(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("config.yml", []byte(`
template_order:
  - post.template.md
generate_readme: false
`), 0644)

	cfg, err := config.Load(mem, "config.yml")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GenerateReadme == nil {
		t.Fatal("expected GenerateReadme to be set, got nil")
	}
	if *cfg.GenerateReadme {
		t.Error("expected GenerateReadme=false, got true")
	}
}

func TestLoad_WithoutValidate(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("no-validate.yml", []byte(`
template_order:
  - post.template.md
`), 0644)

	cfg, err := config.Load(mem, "no-validate.yml")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Validate != nil {
		t.Error("expected nil validate, got non-nil")
	}
}
