package data

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/romayengineer/structured-docs/pkg/structured/fsys"
	"github.com/romayengineer/structured-docs/pkg/structured/schema"
	"gopkg.in/yaml.v3"
)

type DataFile struct {
	SourcePath       string
	TypeName         string
	ExplicitTemplate string
	Fields           map[string]interface{}
}

func LoadAll(fsys fsys.FS, dataDir string, types map[string]*schema.TypeDefinition) ([]*DataFile, error) {
	var files []*DataFile

	err := fsys.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".yml") &&
			!strings.HasSuffix(info.Name(), ".yaml") {
			return nil
		}

		b, err := fsys.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading data file %s: %w", path, err)
		}

		var raw map[string]interface{}
		if err := yaml.Unmarshal(b, &raw); err != nil {
			return fmt.Errorf("parsing data file %s: %w", path, err)
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

		if err := normalizeAndValidate(fields, td); err != nil {
			return fmt.Errorf("data file %s: %w", path, err)
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

func normalizeAndValidate(fields map[string]interface{}, td *schema.TypeDefinition) error {
	for _, fd := range td.Fields {
		val, ok := fields[fd.Name]
		if !ok {
			if fd.Required {
				return fmt.Errorf("missing required field %q", fd.Name)
			}
			continue
		}

		normalized, err := normalizeValue(val, fd.Type)
		if err != nil {
			return fmt.Errorf("field %q: %w", fd.Name, err)
		}
		fields[fd.Name] = normalized
	}

	for name := range fields {
		if !td.HasField(name) {
			return fmt.Errorf("unknown field %q", name)
		}
	}

	return nil
}

func normalizeValue(val interface{}, expected string) (interface{}, error) {
	if val == nil {
		return nil, fmt.Errorf("unexpected nil (expected %s)", expected)
	}

	switch expected {
	case "string":
		switch v := val.(type) {
		case string:
			return v, nil
		case int:
			return fmt.Sprintf("%d", v), nil
		case int32:
			return fmt.Sprintf("%d", v), nil
		case int64:
			return fmt.Sprintf("%d", v), nil
		case float64:
			return fmt.Sprintf("%v", v), nil
		case bool:
			return fmt.Sprintf("%t", v), nil
		default:
			if t, ok := val.(time.Time); ok {
				return t.Format("2006-01-02"), nil
			}
			return nil, fmt.Errorf("expected string, got %T", val)
		}

	case "int":
		switch v := val.(type) {
		case int:
			return v, nil
		case int32:
			return int(v), nil
		case int64:
			return int(v), nil
		case uint:
			return int(v), nil
		case uint32:
			return int(v), nil
		case uint64:
			return int(v), nil
		case float64:
			return int(v), nil
		default:
			return nil, fmt.Errorf("expected int, got %T", val)
		}

	case "float":
		switch v := val.(type) {
		case float64:
			return v, nil
		case float32:
			return float64(v), nil
		case int:
			return float64(v), nil
		case int64:
			return float64(v), nil
		default:
			return nil, fmt.Errorf("expected float, got %T", val)
		}

	case "bool":
		if _, ok := val.(bool); !ok {
			return nil, fmt.Errorf("expected bool, got %T", val)
		}
		return val, nil

	case "[]string":
		list, ok := val.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected []string, got %T", val)
		}
		result := make([]string, len(list))
		for i, item := range list {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("expected string at index %d, got %T", i, item)
			}
			result[i] = s
		}
		return result, nil

	case "[]int":
		list, ok := val.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected []int, got %T", val)
		}
		result := make([]int, len(list))
		for i, item := range list {
			switch v := item.(type) {
			case int:
				result[i] = v
			case int32:
				result[i] = int(v)
			case int64:
				result[i] = int(v)
			case float64:
				result[i] = int(v)
			default:
				return nil, fmt.Errorf("expected int at index %d, got %T", i, item)
			}
		}
		return result, nil

	case "map[string]string":
		m, ok := val.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("expected map[string]string, got %T", val)
		}
		result := make(map[string]string)
		for k, v := range m {
			s, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("expected string value for key %q, got %T", k, v)
			}
			result[k] = s
		}
		return result, nil
	}

	return nil, fmt.Errorf("unsupported type %q", expected)
}
