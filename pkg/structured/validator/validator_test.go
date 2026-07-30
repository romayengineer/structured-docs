package validator_test

import (
	"testing"

	"github.com/romayengineer/structured-docs/pkg/structured/config"
	"github.com/romayengineer/structured-docs/pkg/structured/validator"
)

func rules(h2 int) *config.Validate {
	return &config.Validate{
		LinesBetweenHeaders: &config.HeaderSpacing{
			H2: h2,
		},
	}
}

func fullRules(h1, h2, h3, h4, def int) *config.Validate {
	return &config.Validate{
		LinesBetweenHeaders: &config.HeaderSpacing{
			H1: h1, H2: h2, H3: h3, H4: h4, Default: def,
		},
	}
}

// \n\n = 1 blank line, \n\n\n = 2 blank lines

func TestValidSpacing(t *testing.T) {
	content := "some intro\n\n\n## Section A\ncontent\n\n\n## Section B"
	err := validator.ValidateContent(content, rules(2), "test.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestInvalidSpacing_TooFew(t *testing.T) {
	content := "## Section A\ncontent\n## Section B"
	err := validator.ValidateContent(content, rules(2), "test.md")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "test.md:3: expected 2 blank lines before ## Section B, got 0" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestInvalidSpacing_TooMany(t *testing.T) {
	content := "## Section A\ncontent\n\n\n\n## Section B"
	err := validator.ValidateContent(content, rules(2), "test.md")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "test.md:6: expected 2 blank lines before ## Section B, got 3" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNoHeaders(t *testing.T) {
	content := "just a plain text file\nwith no headers\n"
	err := validator.ValidateContent(content, rules(2), "test.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestNilRules(t *testing.T) {
	err := validator.ValidateContent("## A\n\n## B", nil, "test.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestNilSpacing(t *testing.T) {
	err := validator.ValidateContent("## A\n\n## B", &config.Validate{}, "test.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestZeroSpacing(t *testing.T) {
	r := &config.Validate{
		LinesBetweenHeaders: &config.HeaderSpacing{H2: 0},
	}
	err := validator.ValidateContent("## A\n\n## B", r, "test.md")
	if err != nil {
		t.Fatalf("expected no error (0=disabled), got: %v", err)
	}
}

func TestFirstHeaderSkipped(t *testing.T) {
	content := "## First\n\n\n## Second"
	err := validator.ValidateContent(content, rules(2), "test.md")
	if err != nil {
		t.Fatalf("expected no error for first header (no previous section), got: %v", err)
	}
}

func TestDefaultFallback(t *testing.T) {
	r := &config.Validate{
		LinesBetweenHeaders: &config.HeaderSpacing{Default: 2},
	}
	content := "### A\ncontent\n\n\n### B"
	err := validator.ValidateContent(content, r, "test.md")
	if err != nil {
		t.Fatalf("expected default=2 to apply to h3, got: %v", err)
	}
}

func TestLevelSpecific(t *testing.T) {
	r := fullRules(1, 2, 2, 0, 0)

	content := "# Title\n\n\nintro\n\n\n## Section A\ncontent\n\n\n### Sub A\nmore\n\n\n### Sub B"
	err := validator.ValidateContent(content, r, "test.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestLevelSpecific_Error(t *testing.T) {
	r := fullRules(1, 0, 0, 0, 0)

	content := "# Title\nintro\n# Other"
	err := validator.ValidateContent(content, r, "test.md")
	if err == nil {
		t.Fatal("expected error for h1 spacing, got nil")
	}
	if err.Error() != "test.md:3: expected 1 blank lines before # Other, got 0" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestH5UsesDefault(t *testing.T) {
	r := &config.Validate{
		LinesBetweenHeaders: &config.HeaderSpacing{H2: 2, Default: 1},
	}
	content := "## A\ncontent\n\n##### B\nmore\n\n##### C"
	err := validator.ValidateContent(content, r, "test.md")
	if err != nil {
		t.Fatalf("expected h5 to use default=1, got: %v", err)
	}
}
