package renderer_test

import (
	"testing"

	"github.com/romayengineer/structured-docs/pkg/structured/data"
	"github.com/romayengineer/structured-docs/pkg/structured/renderer"
	"github.com/romayengineer/structured-docs/pkg/structured/resolver"
	"github.com/romayengineer/structured-docs/pkg/structured/template"
)

func TestRender_Success(t *testing.T) {
	job := &resolver.Job{
		Data: &data.DataFile{
			SourcePath: "hello.yml",
			TypeName:   "post",
			Fields:     map[string]interface{}{"title": "Hello World"},
		},
		Template: &template.Template{
			FileName: "post.template.md",
			Content:  "# {{ .title }}",
		},
	}

	out, err := renderer.Render(job)
	if err != nil {
		t.Fatal(err)
	}
	if out != "# Hello World" {
		t.Errorf("expected '# Hello World', got %q", out)
	}
}

func TestRender_ParseError(t *testing.T) {
	job := &resolver.Job{
		Data: &data.DataFile{
			SourcePath: "bad.yml",
			Fields:     map[string]interface{}{},
		},
		Template: &template.Template{
			FileName: "bad.template.md",
			Content:  "{{ .title",
		},
	}

	_, err := renderer.Render(job)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRender_ExecuteError(t *testing.T) {
	job := &resolver.Job{
		Data: &data.DataFile{
			SourcePath: "bad.yml",
			Fields:     map[string]interface{}{"items": "not-a-slice"},
		},
		Template: &template.Template{
			FileName: "bad.template.md",
			Content:  "{{ range .items }}{{ . }}{{ end }}",
		},
	}

	_, err := renderer.Render(job)
	if err == nil {
		t.Fatal("expected error for range over non-iterable, got nil")
	}
}
