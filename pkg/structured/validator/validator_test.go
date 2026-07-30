package validator_test

import (
	"strings"
	"testing"

	"github.com/romayengineer/structured-docs/pkg/structured/config"
	"github.com/romayengineer/structured-docs/pkg/structured/validator"
)

var spacingVal = &validator.HeaderSpacingValidator{}

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
	errs := spacingVal.Validate(content, "test.md", rules(2))
	if len(errs) > 0 {
		t.Fatalf("expected no error, got: %v", errs[0])
	}
}

func TestInvalidSpacing_TooFew(t *testing.T) {
	content := "## Section A\ncontent\n## Section B"
	errs := spacingVal.Validate(content, "test.md", rules(2))
	if len(errs) == 0 {
		t.Fatal("expected error, got nil")
	}
	if errs[0].Error() != "test.md:3: expected 2 blank lines before ## Section B, got 0" {
		t.Errorf("unexpected error: %v", errs[0])
	}
}

func TestInvalidSpacing_TooMany(t *testing.T) {
	content := "## Section A\ncontent\n\n\n\n## Section B"
	errs := spacingVal.Validate(content, "test.md", rules(2))
	if len(errs) == 0 {
		t.Fatal("expected error, got nil")
	}
	if errs[0].Error() != "test.md:6: expected 2 blank lines before ## Section B, got 3" {
		t.Errorf("unexpected error: %v", errs[0])
	}
}

func TestNoHeaders(t *testing.T) {
	content := "just a plain text file\nwith no headers\n"
	errs := spacingVal.Validate(content, "test.md", rules(2))
	if len(errs) > 0 {
		t.Fatalf("expected no error, got: %v", errs[0])
	}
}

func TestNilRules(t *testing.T) {
	errs := spacingVal.Validate("## A\n\n## B", "test.md", nil)
	if len(errs) > 0 {
		t.Fatalf("expected no error, got: %v", errs[0])
	}
}

func TestNilSpacing(t *testing.T) {
	errs := spacingVal.Validate("## A\n\n## B", "test.md", &config.Validate{})
	if len(errs) > 0 {
		t.Fatalf("expected no error, got: %v", errs[0])
	}
}

func TestZeroSpacing(t *testing.T) {
	r := &config.Validate{
		LinesBetweenHeaders: &config.HeaderSpacing{H2: 0},
	}
	errs := spacingVal.Validate("## A\n\n## B", "test.md", r)
	if len(errs) > 0 {
		t.Fatalf("expected no error (0=disabled), got: %v", errs[0])
	}
}

func TestFirstHeaderSkipped(t *testing.T) {
	content := "## First\n\n\n## Second"
	errs := spacingVal.Validate(content, "test.md", rules(2))
	if len(errs) > 0 {
		t.Fatalf("expected no error for first header (no previous section), got: %v", errs[0])
	}
}

func TestDefaultFallback(t *testing.T) {
	r := &config.Validate{
		LinesBetweenHeaders: &config.HeaderSpacing{Default: 2},
	}
	content := "### A\ncontent\n\n\n### B"
	errs := spacingVal.Validate(content, "test.md", r)
	if len(errs) > 0 {
		t.Fatalf("expected default=2 to apply to h3, got: %v", errs[0])
	}
}

func TestLevelSpecific(t *testing.T) {
	r := fullRules(1, 2, 2, 0, 0)
	content := "# Title\n\n\nintro\n\n\n## Section A\ncontent\n\n\n### Sub A\nmore\n\n\n### Sub B"
	errs := spacingVal.Validate(content, "test.md", r)
	if len(errs) > 0 {
		t.Fatalf("expected no error, got: %v", errs[0])
	}
}

func TestLevelSpecific_Error(t *testing.T) {
	r := fullRules(1, 0, 0, 0, 0)
	content := "# Title\nintro\n# Other"
	errs := spacingVal.Validate(content, "test.md", r)
	if len(errs) == 0 {
		t.Fatal("expected error for h1 spacing, got nil")
	}
	if errs[0].Error() != "test.md:3: expected 1 blank lines before # Other, got 0" {
		t.Errorf("unexpected error: %v", errs[0])
	}
}

func TestH5UsesDefault(t *testing.T) {
	r := &config.Validate{
		LinesBetweenHeaders: &config.HeaderSpacing{H2: 2, Default: 1},
	}
	content := "## A\ncontent\n\n##### B\nmore\n\n##### C"
	errs := spacingVal.Validate(content, "test.md", r)
	if len(errs) > 0 {
		t.Fatalf("expected h5 to use default=1, got: %v", errs[0])
	}
}

func TestDefaultValidatorsIncludesHeaderSpacing(t *testing.T) {
	validators := validator.DefaultValidators()
	found := false
	for _, v := range validators {
		if v.Name() == "header-spacing" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("DefaultValidators should include HeaderSpacingValidator")
	}
}

func TestValidContentViaDefault(t *testing.T) {
	content := "## A\ncontent\n\n## B"
	cfg := &config.Validate{
		LinesBetweenHeaders: &config.HeaderSpacing{H2: 1},
	}
	for _, v := range validator.DefaultValidators() {
		errs := v.Validate(content, "test.md", cfg)
		if len(errs) > 0 {
			t.Errorf("validator %s failed: %v", v.Name(), errs[0])
		}
	}
}

func TestDefaultValidatorsSkipOnNilConfig(t *testing.T) {
	for _, v := range validator.DefaultValidators() {
		errs := v.Validate("## A\n## B", "test.md", nil)
		if len(errs) > 0 {
			t.Errorf("validator %s should skip on nil config: %v", v.Name(), errs[0])
		}
	}
}

func TestMultipleErrors(t *testing.T) {
	content := strings.Join([]string{
		"## A",
		"content",
		"## B",
		"more",
		"## C",
	}, "\n")
	errs := spacingVal.Validate(content, "test.md", rules(2))
	if len(errs) != 1 {
		t.Fatalf("expected 1 error (first failure stops), got %d", len(errs))
	}
}
