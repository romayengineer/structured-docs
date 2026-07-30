package data

import (
	"testing"
	"time"

	"github.com/romayengineer/structured-docs/pkg/structured/fsys"
	"github.com/romayengineer/structured-docs/pkg/structured/schema"
)

func TestNormalizeValue_String(t *testing.T) {
	tests := []struct {
		input interface{}
		want  string
	}{
		{"hello", "hello"},
		{int(42), "42"},
		{int32(32), "32"},
		{int64(64), "64"},
		{float64(3.14), "3.14"},
		{true, "true"},
		{false, "false"},
	}
	for _, tc := range tests {
		got, err := normalizeValue(tc.input, "string")
		if err != nil {
			t.Errorf("normalizeValue(%v, string): %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("normalizeValue(%v, string) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestNormalizeValue_String_Time(t *testing.T) {
	val := time.Date(2026, 7, 29, 0, 0, 0, 0, time.UTC)
	got, err := normalizeValue(val, "string")
	if err != nil {
		t.Fatal(err)
	}
	if got != "2026-07-29" {
		t.Errorf("expected 2026-07-29, got %v", got)
	}
}

func TestNormalizeValue_String_UnsupportedType(t *testing.T) {
	_, err := normalizeValue([]int{1, 2, 3}, "string")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNormalizeValue_Int(t *testing.T) {
	tests := []struct {
		input interface{}
		want  int
	}{
		{int(42), 42},
		{int32(32), 32},
		{int64(64), 64},
		{uint(1), 1},
		{uint32(2), 2},
		{uint64(3), 3},
		{float64(99.5), 99},
	}
	for _, tc := range tests {
		got, err := normalizeValue(tc.input, "int")

		if err != nil {
			t.Errorf("normalizeValue(%v, int): %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("normalizeValue(%v, int) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestNormalizeValue_Int_UnsupportedType(t *testing.T) {
	_, err := normalizeValue("hello", "int")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNormalizeValue_Float(t *testing.T) {
	tests := []struct {
		input interface{}
		want  float64
	}{
		{float64(3.14), 3.14},
		{float32(2.5), 2.5},
		{int(10), 10.0},
		{int64(20), 20.0},
	}
	for _, tc := range tests {
		got, err := normalizeValue(tc.input, "float")

		if err != nil {
			t.Errorf("normalizeValue(%v, float): %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("normalizeValue(%v, float) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestNormalizeValue_Float_UnsupportedType(t *testing.T) {
	_, err := normalizeValue(true, "float")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNormalizeValue_Bool(t *testing.T) {
	got, err := normalizeValue(true, "bool")
	if err != nil {
		t.Fatal(err)
	}
	if got != true {
		t.Error("expected true")
	}
}

func TestNormalizeValue_Bool_UnsupportedType(t *testing.T) {
	_, err := normalizeValue("true", "bool")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNormalizeValue_SliceString(t *testing.T) {
	got, err := normalizeValue([]interface{}{"a", "b", "c"}, "[]string")
	if err != nil {
		t.Fatal(err)
	}
	ss, ok := got.([]string)
	if !ok {
		t.Fatalf("expected []string, got %T", got)
	}
	if len(ss) != 3 || ss[0] != "a" || ss[1] != "b" || ss[2] != "c" {
		t.Errorf("expected [a b c], got %v", ss)
	}
}

func TestNormalizeValue_SliceString_WrongType(t *testing.T) {
	_, err := normalizeValue("not a slice", "[]string")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNormalizeValue_SliceString_BadElement(t *testing.T) {
	_, err := normalizeValue([]interface{}{"a", 42}, "[]string")
	if err == nil {
		t.Fatal("expected error for bad element, got nil")
	}
}

func TestNormalizeValue_SliceInt(t *testing.T) {
	got, err := normalizeValue([]interface{}{1, 2, 3}, "[]int")

	if err != nil {
		t.Fatal(err)
	}
	si, ok := got.([]int)
	if !ok {
		t.Fatalf("expected []int, got %T", got)
	}
	if len(si) != 3 || si[0] != 1 || si[1] != 2 || si[2] != 3 {
		t.Errorf("expected [1 2 3], got %v", si)
	}
}

func TestNormalizeValue_SliceInt_WrongType(t *testing.T) {
	_, err := normalizeValue("bad", "[]int")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNormalizeValue_SliceInt_BadElement(t *testing.T) {
	_, err := normalizeValue([]interface{}{"a"}, "[]int")
	if err == nil {
		t.Fatal("expected error for bad element, got nil")
	}
}

func TestNormalizeValue_MapStringString(t *testing.T) {
	got, err := normalizeValue(map[string]interface{}{"key": "value"}, "map[string]string")
	if err != nil {
		t.Fatal(err)
	}
	m, ok := got.(map[string]string)
	if !ok {
		t.Fatalf("expected map[string]string, got %T", got)
	}
	if m["key"] != "value" {
		t.Errorf("expected value, got %q", m["key"])
	}
}

func TestNormalizeValue_MapStringString_WrongType(t *testing.T) {
	_, err := normalizeValue("bad", "map[string]string")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNormalizeValue_MapStringString_BadValue(t *testing.T) {
	_, err := normalizeValue(map[string]interface{}{"key": 42}, "map[string]string")
	if err == nil {
		t.Fatal("expected error for bad value, got nil")
	}
}

func TestNormalizeValue_UnsupportedType(t *testing.T) {
	_, err := normalizeValue("val", "unsupported")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNormalizeValue_Nil(t *testing.T) {
	_, err := normalizeValue(nil, "string")
	if err == nil {
		t.Fatal("expected error for nil, got nil")
	}
}

func TestNormalizeAndValidate_AllFieldsPresent(t *testing.T) {
	fields := map[string]interface{}{"title": "Hello", "count": 42}
	td := &schema.TypeDefinition{
		Fields: []schema.FieldDefinition{
			{Name: "title", Type: "string"},
			{Name: "count", Type: "int"},
		},
	}
	if err := normalizeAndValidate(fields, td); err != nil {
		t.Fatal(err)
	}
	if fields["title"] != "Hello" {
		t.Errorf("expected Hello, got %v", fields["title"])
	}
	if fields["count"] != 42 {
		t.Errorf("expected 42, got %v", fields["count"])
	}
}

func TestNormalizeAndValidate_MissingRequiredField(t *testing.T) {
	fields := map[string]interface{}{}
	td := &schema.TypeDefinition{
		Fields: []schema.FieldDefinition{
			{Name: "title", Type: "string", Required: schema.BoolPtr(true)},
		},
	}
	err := normalizeAndValidate(fields, td)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNormalizeAndValidate_MissingOptionalField(t *testing.T) {
	fields := map[string]interface{}{}
	td := &schema.TypeDefinition{
		Fields: []schema.FieldDefinition{
			{Name: "title", Type: "string", Required: schema.BoolPtr(false)},
		},
	}
	if err := normalizeAndValidate(fields, td); err != nil {
		t.Fatal(err)
	}
}

func TestNormalizeAndValidate_UnknownField(t *testing.T) {
	fields := map[string]interface{}{"unknown": "val"}
	td := &schema.TypeDefinition{
		Fields: []schema.FieldDefinition{
			{Name: "title", Type: "string"},
		},
	}
	err := normalizeAndValidate(fields, td)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNormalizeAndValidate_TypeCoercion(t *testing.T) {
	fields := map[string]interface{}{"count": 42, "ratio": 3.14}
	td := &schema.TypeDefinition{
		Fields: []schema.FieldDefinition{
			{Name: "count", Type: "int"},
			{Name: "ratio", Type: "float"},
		},
	}
	if err := normalizeAndValidate(fields, td); err != nil {
		t.Fatal(err)
	}
	c, ok := fields["count"].(int)
	if !ok || c != 42 {
		t.Errorf("expected int 42, got %T %v", fields["count"], fields["count"])
	}
	r, ok := fields["ratio"].(float64)
	if !ok || r != 3.14 {
		t.Errorf("expected float64 3.14, got %T %v", fields["ratio"], fields["ratio"])
	}
}

func TestNormalizeAndValidate_NormalizeError(t *testing.T) {
	fields := map[string]interface{}{"active": "notbool"}
	td := &schema.TypeDefinition{
		Fields: []schema.FieldDefinition{
			{Name: "active", Type: "bool"},
		},
	}
	err := normalizeAndValidate(fields, td)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoadAll_Success(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("data/blog/hello.yml", []byte(`
type: post
title: Hello
body: World
`), 0644)

	types := map[string]*schema.TypeDefinition{
		"post": {
			Name: "post",
			Fields: []schema.FieldDefinition{
				{Name: "title", Type: "string", Required: schema.BoolPtr(true)},
				{Name: "body", Type: "string"},
			},
		},
	}

	files, err := LoadAll(mem, "data", types)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].SourcePath != "blog/hello.yml" {
		t.Errorf("expected blog/hello.yml, got %q", files[0].SourcePath)
	}
	if files[0].TypeName != "post" {
		t.Errorf("expected post, got %q", files[0].TypeName)
	}
}

func TestLoadAll_SkipsNonYaml(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("data/note.txt", []byte("hello"), 0644)
	mem.WriteFile("data/post.yml", []byte("type: post\ntitle: Hello\n"), 0644)

	types := map[string]*schema.TypeDefinition{
		"post": {Name: "post", Fields: []schema.FieldDefinition{{Name: "title", Type: "string"}}},
	}

	files, err := LoadAll(mem, "data", types)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file (skipping .txt), got %d", len(files))
	}
}

func TestLoadAll_SkipsDirectories(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.MkdirAll("data/subdir", 0755)
	mem.WriteFile("data/subdir/post.yml", []byte("type: post\ntitle: Hello\n"), 0644)

	types := map[string]*schema.TypeDefinition{
		"post": {Name: "post", Fields: []schema.FieldDefinition{{Name: "title", Type: "string"}}},
	}

	files, err := LoadAll(mem, "data", types)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file from subdir, got %d", len(files))
	}
}

func TestLoadAll_ParseError(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("data/post.yml", []byte(`{{{invalid`), 0644)

	types := map[string]*schema.TypeDefinition{
		"post": {Name: "post", Fields: []schema.FieldDefinition{{Name: "title", Type: "string"}}},
	}

	_, err := LoadAll(mem, "data", types)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoadAll_MissingType(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("data/post.yml", []byte(`title: Hello`), 0644)

	types := map[string]*schema.TypeDefinition{}

	_, err := LoadAll(mem, "data", types)
	if err == nil {
		t.Fatal("expected error for missing type, got nil")
	}
}

func TestLoadAll_TypeNotString(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("data/post.yml", []byte(`type: 123`), 0644)

	types := map[string]*schema.TypeDefinition{}

	_, err := LoadAll(mem, "data", types)
	if err == nil {
		t.Fatal("expected error for non-string type, got nil")
	}
}

func TestLoadAll_UnknownType(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("data/post.yml", []byte(`type: unknown`), 0644)

	types := map[string]*schema.TypeDefinition{}

	_, err := LoadAll(mem, "data", types)
	if err == nil {
		t.Fatal("expected error for unknown type, got nil")
	}
}

func TestLoadAll_ExplicitTemplate(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("data/post.yml", []byte(`
type: post
template: custom.template.md
title: Hello
`), 0644)

	types := map[string]*schema.TypeDefinition{
		"post": {Name: "post", Fields: []schema.FieldDefinition{{Name: "title", Type: "string"}}},
	}

	files, err := LoadAll(mem, "data", types)
	if err != nil {
		t.Fatal(err)
	}
	if files[0].ExplicitTemplate != "custom.template.md" {
		t.Errorf("expected custom template, got %q", files[0].ExplicitTemplate)
	}
}

func TestLoadAll_YamlExtension(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("data/page.yaml", []byte(`
type: page
title: Hello
`), 0644)

	types := map[string]*schema.TypeDefinition{
		"page": {Name: "page", Fields: []schema.FieldDefinition{{Name: "title", Type: "string"}}},
	}

	files, err := LoadAll(mem, "data", types)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].SourcePath != "page.yaml" {
		t.Errorf("expected page.yaml, got %q", files[0].SourcePath)
	}
}

func TestLoadAll_MissingRequiredField(t *testing.T) {
	mem := fsys.NewMemFS()
	mem.WriteFile("data/post.yml", []byte(`
type: post
`), 0644)

	types := map[string]*schema.TypeDefinition{
		"post": {Name: "post", Fields: []schema.FieldDefinition{
			{Name: "title", Type: "string", Required: schema.BoolPtr(true)},
		}},
	}

	_, err := LoadAll(mem, "data", types)
	if err == nil {
		t.Fatal("expected error for missing required field, got nil")
	}
}
