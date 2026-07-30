package validator

import (
	"fmt"
	"strings"

	"github.com/romayengineer/structured-docs/pkg/structured/config"
)

func ValidateContent(content string, rules *config.Validate, filePath string) error {
	if rules == nil || rules.LinesBetweenHeaders == nil {
		return nil
	}

	spacing := rules.LinesBetweenHeaders

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		level := headerLevel(line)
		if level == 0 {
			continue
		}

		expected := spacingForLevel(spacing, level)
		if expected <= 0 {
			continue
		}

		if i == 0 {
			continue
		}

		blankLines := countBlankLinesBefore(lines, i)
		if blankLines != expected {
			headerText := strings.TrimSpace(line)
			return fmt.Errorf("%s:%d: expected %d blank lines before %s, got %d",
				filePath, i+1, expected, headerText, blankLines)
		}
	}

	return nil
}

func headerLevel(line string) int {
	if line == "" {
		return 0
	}
	for i, c := range line {
		if c != '#' {
			if i > 0 && i <= 6 && i < len(line) && line[i] == ' ' {
				return i
			}
			return 0
		}
	}
	return 0
}

func spacingForLevel(s *config.HeaderSpacing, level int) int {
	switch level {
	case 1:
		return pick(s.H1, s.Default)
	case 2:
		return pick(s.H2, s.Default)
	case 3:
		return pick(s.H3, s.Default)
	case 4:
		return pick(s.H4, s.Default)
	default:
		return s.Default
	}
}

func pick(levelVal, defaultVal int) int {
	if levelVal > 0 {
		return levelVal
	}
	return defaultVal
}

func countBlankLinesBefore(lines []string, idx int) int {
	count := 0
	for j := idx - 1; j >= 0; j-- {
		if strings.TrimSpace(lines[j]) == "" {
			count++
		} else {
			break
		}
	}
	return count
}
