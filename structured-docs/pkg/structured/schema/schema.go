package schema

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/romayengineer/structured-docs/pkg/structured/fsys"
	"gopkg.in/yaml.v3"
)

type FieldDefinition struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Required bool   `yaml:"required"`
}

type TypeDefinition struct {
	Name        string
	Description string            `yaml:"description"`
	Fields      []FieldDefinition `yaml:"fields"`
}

func (t *TypeDefinition) FieldNames() []string {
	names := make([]string, len(t.Fields))
	for i, f := range t.Fields {
		names[i] = f.Name
	}
	return names
}

func (t *TypeDefinition) HasField(name string) bool {
	for _, f := range t.Fields {
		if f.Name == name {
			return true
		}
	}
	return false
}

func (t *TypeDefinition) Field(name string) *FieldDefinition {
	for _, f := range t.Fields {
		if f.Name == name {
			return &f
		}
	}
	return nil
}

func LoadAll(fsys fsys.FS, schemaDir string) (map[string]*TypeDefinition, error) {
	entries, err := fsys.ReadDir(schemaDir)
	if err != nil {
		return nil, fmt.Errorf("reading schema dir %s: %w", schemaDir, err)
	}

	types := make(map[string]*TypeDefinition)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".yml") && !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		typeName := strings.TrimSuffix(entry.Name(), ".yml")
		typeName = strings.TrimSuffix(typeName, ".yaml")

		path := filepath.Join(schemaDir, entry.Name())
		b, err := fsys.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading schema %s: %w", path, err)
		}

		var td TypeDefinition
		if err := yaml.Unmarshal(b, &td); err != nil {
			return nil, fmt.Errorf("parsing schema %s: %w", path, err)
		}

		td.Name = typeName

		for _, f := range td.Fields {
			if !isValidFieldType(f.Type) {
				return nil, fmt.Errorf("schema %s: invalid field type %q for field %q", path, f.Type, f.Name)
			}
		}

		types[typeName] = &td
	}

	return types, nil
}

func isValidFieldType(t string) bool {
	switch t {
	case "string", "int", "float", "bool",
		"[]string", "[]int",
		"map[string]string":
		return true
	}
	return false
}
