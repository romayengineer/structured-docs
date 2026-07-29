package data

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/romayengineer/structured-docs/pkg/structured/schema"
	"gopkg.in/yaml.v3"
)

type DataFile struct {
	SourcePath string
	TypeName   string
	ExplicitTemplate string
	Fields     map[string]interface{}
}

func loadRaw(path string) (map[string]interface{}, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading data file %s: %w", path, err)
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("parsing data file %s: %w", path, err)
	}

	return raw, nil
}

func LoadAll(dataDir string, types map[string]*schema.TypeDefinition) ([]*DataFile, error) {
	var files []*DataFile

	err := filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".yml") && !strings.HasSuffix(info.Name(), ".yaml") {
			return nil
		}

		raw, err := loadRaw(path)
		if err != nil {
			return fmt.Errorf("loading data: %w", err)
		}

		typeVal, ok := raw["type"]
		if !ok {
			return fmt.Errorf("data file %s: missing 'type' field", path)
		}

		typeName, ok := typeVal.(string)
		if !ok {
			return fmt.Errorf("data file %s: 'type' field must be a string", path)
		}

		td, exists := types[typeName]
		if !exists {
			return fmt.Errorf("data file %s: unknown type %q", path, typeName)
		}

		explicitTemplate, _ := raw["template"].(string)

		fields := make(map[string]interface{})
		for k, v := range raw {
			if k == "type" || k == "template" {
				continue
			}
			fields[k] = v
		}

		if err := validate(fields, td); err != nil {
			return fmt.Errorf("data file %s: validation error: %w", path, err)
		}

		relPath, err := filepath.Rel(dataDir, path)
		if err != nil {
			return fmt.Errorf("getting relative path for %s: %w", path, err)
		}

		files = append(files, &DataFile{
			SourcePath:       relPath,
			TypeName:         typeName,
			ExplicitTemplate: explicitTemplate,
			Fields:           fields,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking data dir: %w", err)
	}

	return files, nil
}

func validate(fields map[string]interface{}, td *schema.TypeDefinition) error {
	for _, f := range td.Fields {
		if f.Required {
			val, ok := fields[f.Name]
			if !ok {
				return fmt.Errorf("missing required field %q", f.Name)
			}
			if err := checkType(val, f.Type); err != nil {
				return fmt.Errorf("field %q: %w", f.Name, err)
			}
		}
	}

	for name, val := range fields {
		if !td.HasField(name) {
			return fmt.Errorf("unknown field %q", name)
		}
		fd := td.Field(name)
		if err := checkType(val, fd.Type); err != nil {
			return fmt.Errorf("field %q: %w", name, err)
		}
	}

	return nil
}

func checkType(val interface{}, expected string) error {
	if val == nil {
		return fmt.Errorf("unexpected nil value (expected %s)", expected)
	}

	switch expected {
	case "string":
		if _, ok := val.(string); !ok {
			return fmt.Errorf("expected string, got %T", val)
		}
	case "int":
		switch val.(type) {
		case int, int32, int64, uint, uint32, uint64:
		default:
			return fmt.Errorf("expected int, got %T", val)
		}
	case "float":
		switch val.(type) {
		case float32, float64:
		case int, int32, int64:
		default:
			return fmt.Errorf("expected float, got %T", val)
		}
	case "bool":
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("expected bool, got %T", val)
		}
	case "[]string":
		list, ok := val.([]interface{})
		if !ok {
			return fmt.Errorf("expected []string, got %T", val)
		}
		for i, item := range list {
			if _, ok := item.(string); !ok {
				return fmt.Errorf("expected string at index %d, got %T", i, item)
			}
		}
	case "[]int":
		list, ok := val.([]interface{})
		if !ok {
			return fmt.Errorf("expected []int, got %T", val)
		}
		for i, item := range list {
			switch item.(type) {
			case int, int32, int64, uint, uint32, uint64:
			default:
				return fmt.Errorf("expected int at index %d, got %T", i, item)
			}
		}
	case "map[string]string":
		m, ok := val.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected map[string]string, got %T", val)
		}
		for k, v := range m {
			if _, ok := v.(string); !ok {
				return fmt.Errorf("expected string value for key %q, got %T", k, v)
			}
		}
	}

	return nil
}
