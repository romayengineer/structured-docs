package schema

import (
	"testing"

	"github.com/romayengineer/structured-docs/pkg/structured/fsys"
)

func TestFieldNames(t *testing.T) {
	td := &TypeDefinition{
		Name: "post",
		Fields: []FieldDefinition{
			{Name: "title", Type: "string"},
			{Name: "body", Type: "string"},
		},
	}
	names := td.FieldNames()
	if len(names) != 2 || names[0] != "title" || names[1] != "body" {
		t.Errorf("expected [title body], got %v", names)
	}
}

func TestFieldNames_Empty(t *testing.T) {
	td := &TypeDefinition{Name: "empty"}
	names := td.FieldNames()
	if len(names) != 0 {
		t.Errorf("expected empty, got %v", names)
	}
}

func TestHasField(t *testing.T) {
	td := &TypeDefinition{
		Fields: []FieldDefinition{
			{Name: "title", Type: "string"},
		},
	}
	if !td.HasField("title") {
		t.Error("expected HasField(title) to be true")
	}
	if td.HasField("nonexistent") {
		t.Error("expected HasField(nonexistent) to be false")
	}
}

func TestField_Found(t *testing.T) {
	td := &TypeDefinition{
		Fields: []FieldDefinition{
			{Name: "title", Type: "string", Required: BoolPtr(true)},
		},
	}
	f := td.Field("title")
	if f == nil {
		t.Fatal("expected non-nil field")
	}
	if f.Name != "title" || f.Type != "string" || !f.IsRequired() {
		t.Error("field attributes mismatch")
	}
}

func TestField_NotFound(t *testing.T) {
	td := &TypeDefinition{}
	f := td.Field("missing")
	if f != nil {
		t.Error("expected nil for missing field")
	}
}

func TestIsValidFieldType(t *testing.T) {
	valid := []string{"string", "int", "float", "bool", "[]string", "[]int", "map[string]string"}
	for _, typ := range valid {
		if !isValidFieldType(typ) {
			t.Errorf("expected %q to be valid", typ)
		}
	}
}

func TestIsValidFieldType_Invalid(t *testing.T) {
	invalid := []string{"", "uint", "[]float", "map[int]string", "any", "number"}
	for _, typ := range invalid {
		if isValidFieldType(typ) {
			t.Errorf("expected %q to be invalid", typ)
		}
	}
}

func TestLoadAll_Success(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("schema/post.yml", []byte(`
description: A blog post
fields:
  - name: title
    type: string
    required: true
  - name: body
    type: string
`), 0644)

	types, err := LoadAll(mem, "schema")
	if err != nil {
		t.Fatal(err)
	}
	if len(types) != 1 {
		t.Fatalf("expected 1 type, got %d", len(types))
	}
	td := types["post"]
	if td == nil {
		t.Fatal("expected type 'post'")
	}
	if td.Name != "post" {
		t.Errorf("expected name 'post', got %q", td.Name)
	}
	if td.Description != "A blog post" {
		t.Errorf("expected description 'A blog post', got %q", td.Description)
	}
	if len(td.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(td.Fields))
	}
}

func TestLoadAll_YamlExtension(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("schema/page.yaml", []byte(`
fields:
  - name: title
    type: string
`), 0644)

	types, err := LoadAll(mem, "schema")
	if err != nil {
		t.Fatal(err)
	}
	if types["page"] == nil {
		t.Error("expected type 'page' from .yaml file")
	}
}

func TestLoadAll_ReadDirError(t *testing.T) {
	mem := fsys.NewMemFS()
	_, err := LoadAll(mem, "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoadAll_ParseError(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("schema/post.yml", []byte(`{{{invalid`), 0644)

	_, err := LoadAll(mem, "schema")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoadAll_InvalidFieldType(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("schema/post.yml", []byte(`
fields:
  - name: count
    type: uint
`), 0644)

	_, err := LoadAll(mem, "schema")
	if err == nil {
		t.Fatal("expected error for invalid field type, got nil")
	}
}

func TestLoadAll_SkipsDirectoriesAndNonYaml(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.MkdirAll("schema/subdir", 0755)
	mem.WriteFile("schema/ignored.txt", []byte("hello"), 0644)
	mem.WriteFile("schema/post.yml", []byte(`
fields:
  - name: title
    type: string
`), 0644)

	types, err := LoadAll(mem, "schema")
	if err != nil {
		t.Fatal(err)
	}
	if len(types) != 1 {
		t.Errorf("expected 1 type, got %d", len(types))
	}
}
