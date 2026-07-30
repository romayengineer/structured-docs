package template

import (
	"testing"

	"github.com/romayengineer/structured-docs/pkg/structured/fsys"
)

func TestOutputExt_Markdown(t *testing.T) {
	if got := outputExt("post.template.md"); got != ".md" {
		t.Errorf("expected .md, got %q", got)
	}
}

func TestOutputExt_HTML(t *testing.T) {
	if got := outputExt("page.template.html"); got != ".html" {
		t.Errorf("expected .html, got %q", got)
	}
}

func TestExtractFields_Basic(t *testing.T) {
	content := `# {{ .title }}
{{ .body }}
{{ range .tags }} {{ . }}{{ end }}`
	fields, err := extractFields(content)
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 3 {
		t.Fatalf("expected 3 fields, got %v", fields)
	}
	expect := map[string]bool{"title": true, "body": true, "tags": true}
	for _, f := range fields {
		if !expect[f] {
			t.Errorf("unexpected field %q", f)
		}
		delete(expect, f)
	}
}

func TestExtractFields_IfWithElse(t *testing.T) {
	content := `{{ if .show }}{{ .title }}{{ else }}{{ .fallback }}{{ end }}`
	fields, err := extractFields(content)
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 3 {
		t.Fatalf("expected 3 fields, got %v", fields)
	}
}

func TestExtractFields_WithNode(t *testing.T) {
	content := `{{ with .author }}{{ .name }}{{ end }}`
	fields, err := extractFields(content)
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 2 {
		t.Fatalf("expected 2 fields (author, name), got %v", fields)
	}
}

func TestExtractFields_ChainedField(t *testing.T) {
	content := `{{ .author.name }}`
	fields, err := extractFields(content)
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 1 || fields[0] != "author" {
		t.Errorf("expected only 'author' (top-level), got %v", fields)
	}
}

func TestExtractFields_TemplateNode(t *testing.T) {
	content := `{{ template "header" .data }}`
	fields, err := extractFields(content)
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 1 || fields[0] != "data" {
		t.Errorf("expected [data], got %v", fields)
	}
}

func TestExtractFields_VariableDeclaration(t *testing.T) {
	content := `{{ range $i, $v := .items }}{{ $i }}{{ end }}`
	fields, err := extractFields(content)
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 1 || fields[0] != "items" {
		t.Errorf("expected [items], got %v", fields)
	}
}

func TestExtractFields_Empty(t *testing.T) {
	fields, err := extractFields("")
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 0 {
		t.Errorf("expected no fields, got %v", fields)
	}
}

func TestExtractFields_NoFields(t *testing.T) {
	fields, err := extractFields("static content")
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 0 {
		t.Errorf("expected no fields, got %v", fields)
	}
}

func TestExtractFields_ParseError(t *testing.T) {
	_, err := extractFields("{{ .title")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExtractFields_Deduplicates(t *testing.T) {
	content := `{{ .title }} {{ .title }}`
	fields, err := extractFields(content)
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 1 {
		t.Errorf("expected 1 unique field, got %v", fields)
	}
}

func TestLoadAll_Success(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("templates/post.template.md", []byte(`
# {{ .title }}
{{ .body }}
`), 0644)

	templates, err := LoadAll(mem, "templates")
	if err != nil {
		t.Fatal(err)
	}
	if len(templates) != 1 {
		t.Fatalf("expected 1 template, got %d", len(templates))
	}
	if templates[0].FileName != "post.template.md" {
		t.Errorf("expected post.template.md, got %q", templates[0].FileName)
	}
	if templates[0].OutputExt != ".md" {
		t.Errorf("expected .md, got %q", templates[0].OutputExt)
	}
	if len(templates[0].RequiredFields) != 2 {
		t.Errorf("expected 2 required fields, got %v", templates[0].RequiredFields)
	}
}

func TestLoadAll_HTMLTemplate(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("templates/page.template.html", []byte(`<h1>{{ .title }}</h1>`), 0644)

	templates, err := LoadAll(mem, "templates")
	if err != nil {
		t.Fatal(err)
	}
	if len(templates) != 1 {
		t.Fatalf("expected 1 template, got %d", len(templates))
	}
	if templates[0].OutputExt != ".html" {
		t.Errorf("expected .html output ext, got %q", templates[0].OutputExt)
	}
}

func TestLoadAll_SkipsNonTemplateFiles(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("templates/note.txt", []byte("hello"), 0644)
	mem.WriteFile("templates/post.template.md", []byte(`{{ .title }}`), 0644)

	templates, err := LoadAll(mem, "templates")
	if err != nil {
		t.Fatal(err)
	}
	if len(templates) != 1 {
		t.Errorf("expected 1 template, got %d", len(templates))
	}
}

func TestLoadAll_SkipsDirectories(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.MkdirAll("templates/subdir", 0755)
	mem.WriteFile("templates/subdir/post.template.md", []byte(`{{ .title }}`), 0644)

	templates, err := LoadAll(mem, "templates")
	if err != nil {
		t.Fatal(err)
	}
	if len(templates) != 1 {
		t.Errorf("expected 1 template from subdir, got %d", len(templates))
	}
}

func TestLoadAll_ExtractFieldsError(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("templates/bad.template.md", []byte(`{{ .title`), 0644)

	_, err := LoadAll(mem, "templates")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoadAll_SubdirectoryTemplate(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.MkdirAll("templates/blog", 0755)
	mem.WriteFile("templates/blog/post.template.md", []byte(`{{ .title }}`), 0644)

	templates, err := LoadAll(mem, "templates")
	if err != nil {
		t.Fatal(err)
	}
	if len(templates) != 1 {
		t.Fatalf("expected 1 template, got %d", len(templates))
	}
	if templates[0].FileName != "blog/post.template.md" {
		t.Errorf("expected blog/post.template.md, got %q", templates[0].FileName)
	}
}

func TestExtractFields_RangeElse(t *testing.T) {
	content := `{{ range .items }}{{ . }}{{ else }}{{ .fallback }}{{ end }}`
	fields, err := extractFields(content)
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 2 {
		t.Fatalf("expected 2 fields (items, fallback), got %v", fields)
	}
}

func TestExtractFields_WithElse(t *testing.T) {
	content := `{{ with .author }}{{ .name }}{{ else }}{{ .fallback }}{{ end }}`
	fields, err := extractFields(content)
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 3 {
		t.Fatalf("expected 3 fields (author, name, fallback), got %v", fields)
	}
}
