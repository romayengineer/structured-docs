package resolver

import (
	"testing"

	"github.com/romayengineer/structured-docs/pkg/structured/data"
	"github.com/romayengineer/structured-docs/pkg/structured/schema"
	"github.com/romayengineer/structured-docs/pkg/structured/template"
)

func TestFieldsCompatible(t *testing.T) {
	types := map[string]*schema.TypeDefinition{
		"post": {
			Name: "post",
			Fields: []schema.FieldDefinition{
				{Name: "title", Type: "string"},
				{Name: "body", Type: "string"},
			},
		},
	}

	tmpl := &template.Template{
		FileName:       "post.template.md",
		RequiredFields: []string{"title", "body"},
	}

	df := &data.DataFile{TypeName: "post"}

	if !fieldsCompatible(tmpl, df, types) {
		t.Error("expected compatible")
	}
}

func TestFieldsCompatible_Incompatible(t *testing.T) {
	types := map[string]*schema.TypeDefinition{
		"post": {
			Name: "post",
			Fields: []schema.FieldDefinition{
				{Name: "title", Type: "string"},
			},
		},
	}

	tmpl := &template.Template{
		RequiredFields: []string{"title", "nonexistent"},
	}

	df := &data.DataFile{TypeName: "post"}

	if fieldsCompatible(tmpl, df, types) {
		t.Error("expected incompatible")
	}
}

func TestFieldsCompatible_UnknownType(t *testing.T) {
	types := map[string]*schema.TypeDefinition{}
	tmpl := &template.Template{RequiredFields: []string{"title"}}
	df := &data.DataFile{TypeName: "unknown"}

	if fieldsCompatible(tmpl, df, types) {
		t.Error("expected incompatible for unknown type")
	}
}

func TestResolveAll_ExplicitTemplate(t *testing.T) {
	dataFiles := []*data.DataFile{
		{SourcePath: "hello.yml", TypeName: "post", ExplicitTemplate: "custom.template.md"},
	}
	templates := []*template.Template{
		{FileName: "custom.template.md", RequiredFields: []string{"title"}},
	}
	types := map[string]*schema.TypeDefinition{
		"post": {Name: "post", Fields: []schema.FieldDefinition{{Name: "title", Type: "string"}}},
	}

	jobs, err := ResolveAll(dataFiles, templates, nil, types)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Template.FileName != "custom.template.md" {
		t.Errorf("expected custom.template.md, got %q", jobs[0].Template.FileName)
	}
}

func TestResolveAll_ExplicitTemplateNotFound(t *testing.T) {
	dataFiles := []*data.DataFile{
		{SourcePath: "hello.yml", TypeName: "post", ExplicitTemplate: "missing.template.md"},
	}
	templates := []*template.Template{}
	types := map[string]*schema.TypeDefinition{}

	_, err := ResolveAll(dataFiles, templates, nil, types)
	if err == nil {
		t.Fatal("expected error for missing template, got nil")
	}
}

func TestResolveAll_ExplicitTemplateIncompatible(t *testing.T) {
	dataFiles := []*data.DataFile{
		{SourcePath: "hello.yml", TypeName: "post", ExplicitTemplate: "strict.template.md"},
	}
	templates := []*template.Template{
		{FileName: "strict.template.md", RequiredFields: []string{"missing_field"}},
	}
	types := map[string]*schema.TypeDefinition{
		"post": {Name: "post", Fields: []schema.FieldDefinition{{Name: "title", Type: "string"}}},
	}

	_, err := ResolveAll(dataFiles, templates, nil, types)
	if err == nil {
		t.Fatal("expected error for incompatible template, got nil")
	}
}

func TestResolveAll_ImplicitTemplate(t *testing.T) {
	dataFiles := []*data.DataFile{
		{SourcePath: "hello.yml", TypeName: "post"},
	}
	templates := []*template.Template{
		{FileName: "a.template.md", RequiredFields: []string{"title"}},
		{FileName: "b.template.md", RequiredFields: []string{"title"}},
	}
	types := map[string]*schema.TypeDefinition{
		"post": {Name: "post", Fields: []schema.FieldDefinition{{Name: "title", Type: "string"}}},
	}

	jobs, err := ResolveAll(dataFiles, templates, []string{"a.template.md", "b.template.md"}, types)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Template.FileName != "a.template.md" {
		t.Errorf("expected first matching template a.template.md, got %q", jobs[0].Template.FileName)
	}
}

func TestResolveAll_ImplicitTemplateSkipsMissing(t *testing.T) {
	dataFiles := []*data.DataFile{
		{SourcePath: "hello.yml", TypeName: "post"},
	}
	templates := []*template.Template{
		{FileName: "b.template.md", RequiredFields: []string{"title"}},
	}
	types := map[string]*schema.TypeDefinition{
		"post": {Name: "post", Fields: []schema.FieldDefinition{{Name: "title", Type: "string"}}},
	}

	jobs, err := ResolveAll(dataFiles, templates, []string{"missing.template.md", "b.template.md"}, types)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
}

func TestResolveAll_NoCompatibleTemplate(t *testing.T) {
	dataFiles := []*data.DataFile{
		{SourcePath: "hello.yml", TypeName: "post"},
	}
	templates := []*template.Template{
		{FileName: "post.template.md", RequiredFields: []string{"nonexistent"}},
	}
	types := map[string]*schema.TypeDefinition{
		"post": {Name: "post", Fields: []schema.FieldDefinition{{Name: "title", Type: "string"}}},
	}

	_, err := ResolveAll(dataFiles, templates, []string{"post.template.md"}, types)
	if err == nil {
		t.Fatal("expected error for no compatible template, got nil")
	}
}

func TestResolveAll_MultipleDataFiles(t *testing.T) {
	dataFiles := []*data.DataFile{
		{SourcePath: "a.yml", TypeName: "post"},
		{SourcePath: "b.yml", TypeName: "post"},
	}
	templates := []*template.Template{
		{FileName: "post.template.md", RequiredFields: []string{"title"}},
	}
	types := map[string]*schema.TypeDefinition{
		"post": {Name: "post", Fields: []schema.FieldDefinition{{Name: "title", Type: "string"}}},
	}

	jobs, err := ResolveAll(dataFiles, templates, []string{"post.template.md"}, types)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
}
