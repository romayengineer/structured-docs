package validator

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/romayengineer/structured-docs/pkg/structured/config"
)

type FileNameValidator struct{}

func (v *FileNameValidator) Name() string { return "file-name" }

func (v *FileNameValidator) Validate(content string, filePath string, cfg *config.Validate) []error {
	if cfg == nil || cfg.FileName == nil {
		return nil
	}

	rule := cfg.FileName
	pattern := rule.Pattern

	if pattern == "" {
		switch rule.Style {
		case "kebab":
			pattern = `^[a-z0-9]+(-[a-z0-9]+)*$`
		case "snake":
			pattern = `^[a-z0-9]+(_[a-z0-9]+)*$`
		case "camel":
			pattern = `^[a-z]+([A-Z][a-z0-9]*)*$`
		case "lowercase":
			pattern = `^[a-z0-9]+$`
		default:
			return nil
		}
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return []error{fmt.Errorf("%s: invalid file_name pattern %q: %w", filePath, pattern, err)}
	}

	base := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	if !re.MatchString(base) {
		desc := rule.Style
		if desc == "" {
			desc = fmt.Sprintf("pattern %s", pattern)
		}
		return []error{fmt.Errorf("%s: file name %q does not match %s style", filePath, base, desc)}
	}

	return nil
}
