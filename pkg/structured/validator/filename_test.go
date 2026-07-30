package validator_test

import (
	"strings"
	"testing"

	"github.com/romayengineer/structured-docs/pkg/structured/config"
	"github.com/romayengineer/structured-docs/pkg/structured/validator"
)

var fnVal = &validator.FileNameValidator{}

func fileRule(style string) *config.Validate {
	return &config.Validate{
		FileName: &config.FileNameRule{Style: style},
	}
}

func TestFileName_Kebab_Valid(t *testing.T) {
	errs := fnVal.Validate("content", "output/blog/go-generics.md", fileRule("kebab"))
	if len(errs) > 0 {
		t.Fatalf("expected no error, got: %v", errs[0])
	}
}

func TestFileName_Kebab_Invalid(t *testing.T) {
	errs := fnVal.Validate("content", "output/blog/GoGenerics.md", fileRule("kebab"))
	if len(errs) == 0 {
		t.Fatal("expected error, got nil")
	}
	if errs[0].Error() != `output/blog/GoGenerics.md: file name "GoGenerics" does not match kebab style` {
		t.Errorf("unexpected error: %v", errs[0])
	}
}

func TestFileName_Snake_Valid(t *testing.T) {
	errs := fnVal.Validate("content", "output/blog/go_generics.md", fileRule("snake"))
	if len(errs) > 0 {
		t.Fatalf("expected no error, got: %v", errs[0])
	}
}

func TestFileName_Snake_Invalid(t *testing.T) {
	errs := fnVal.Validate("content", "output/blog/go-generics.md", fileRule("snake"))
	if len(errs) == 0 {
		t.Fatal("expected error, got nil")
	}
}

func TestFileName_Camel_Valid(t *testing.T) {
	errs := fnVal.Validate("content", "output/blog/goGenerics.md", fileRule("camel"))
	if len(errs) > 0 {
		t.Fatalf("expected no error, got: %v", errs[0])
	}
}

func TestFileName_Camel_Invalid(t *testing.T) {
	errs := fnVal.Validate("content", "output/blog/go-generics.md", fileRule("camel"))
	if len(errs) == 0 {
		t.Fatal("expected error, got nil")
	}
}

func TestFileName_Lowercase_Valid(t *testing.T) {
	errs := fnVal.Validate("content", "output/blog/gogenerics.md", fileRule("lowercase"))
	if len(errs) > 0 {
		t.Fatalf("expected no error, got: %v", errs[0])
	}
}

func TestFileName_Lowercase_Invalid(t *testing.T) {
	errs := fnVal.Validate("content", "output/blog/go-generics.md", fileRule("lowercase"))
	if len(errs) == 0 {
		t.Fatal("expected error, got nil")
	}
}

func TestFileName_NilRule(t *testing.T) {
	errs := fnVal.Validate("content", "output/blog/ANY-name.md", &config.Validate{})
	if len(errs) > 0 {
		t.Fatalf("expected no error, got: %v", errs[0])
	}
}

func TestFileName_NilValidate(t *testing.T) {
	errs := fnVal.Validate("content", "output/blog/ANY-name.md", nil)
	if len(errs) > 0 {
		t.Fatalf("expected no error, got: %v", errs[0])
	}
}

func TestFileName_EmptyStyle(t *testing.T) {
	r := &config.Validate{FileName: &config.FileNameRule{Style: ""}}
	errs := fnVal.Validate("content", "output/blog/ANY-name.md", r)
	if len(errs) > 0 {
		t.Fatalf("expected no error (empty style), got: %v", errs[0])
	}
}

func TestFileName_CustomPattern(t *testing.T) {
	r := &config.Validate{
		FileName: &config.FileNameRule{Pattern: `^[a-z]+(-[a-z]+)*$`},
	}
	errs := fnVal.Validate("content", "output/blog/my-post.md", r)
	if len(errs) > 0 {
		t.Fatalf("expected no error, got: %v", errs[0])
	}
}

func TestFileName_CustomPattern_Fail(t *testing.T) {
	r := &config.Validate{
		FileName: &config.FileNameRule{Pattern: `^[a-z]+(-[a-z]+)*$`},
	}
	errs := fnVal.Validate("content", "output/blog/123.md", r)
	if len(errs) == 0 {
		t.Fatal("expected error, got nil")
	}
}

func TestFileName_InvalidPattern(t *testing.T) {
	r := &config.Validate{
		FileName: &config.FileNameRule{Pattern: `[invalid`},
	}
	errs := fnVal.Validate("content", "output/blog/x.md", r)
	if len(errs) == 0 {
		t.Fatal("expected error for invalid regex, got nil")
	}
}

func boolPtr(b bool) *bool { return &b }

func TestFileName_LeadingDigitUnsetAllowsDigits(t *testing.T) {
	r := &config.Validate{
		FileName: &config.FileNameRule{Style: "kebab"},
	}
	errs := fnVal.Validate("content", "output/blog/123-post.md", r)
	if len(errs) > 0 {
		t.Fatalf("expected no error (default allows leading digits), got: %v", errs[0])
	}
}

func TestFileName_LeadingDigitExplicitTrueAllows(t *testing.T) {
	r := &config.Validate{
		FileName: &config.FileNameRule{
			Style:             "kebab",
			AllowLeadingDigit: boolPtr(true),
		},
	}
	errs := fnVal.Validate("content", "output/blog/123-post.md", r)
	if len(errs) > 0 {
		t.Fatalf("expected no error (explicit true allows digits), got: %v", errs[0])
	}
}

func TestFileName_LeadingDigitFalseRejects(t *testing.T) {
	r := &config.Validate{
		FileName: &config.FileNameRule{
			Style:             "kebab",
			AllowLeadingDigit: boolPtr(false),
		},
	}
	errs := fnVal.Validate("content", "output/blog/123-post.md", r)
	if len(errs) == 0 {
		t.Fatal("expected error for leading digit, got nil")
	}
	if errs[0].Error() != `output/blog/123-post.md: file name "123-post" must not start with a digit` {
		t.Errorf("unexpected error: %v", errs[0])
	}
}

func TestFileName_LeadingDigitFalseValidName(t *testing.T) {
	r := &config.Validate{
		FileName: &config.FileNameRule{
			Style:             "kebab",
			AllowLeadingDigit: boolPtr(false),
		},
	}
	errs := fnVal.Validate("content", "output/blog/go-post.md", r)
	if len(errs) > 0 {
		t.Fatalf("expected no error (valid name with leading digit false), got: %v", errs[0])
	}
}

func TestFileName_LeadingDigitFalseWithPattern(t *testing.T) {
	r := &config.Validate{
		FileName: &config.FileNameRule{
			Pattern:           `^[a-zA-Z0-9]+$`,
			AllowLeadingDigit: boolPtr(false),
		},
	}
	errs := fnVal.Validate("content", "output/blog/abc123.md", r)
	if len(errs) > 0 {
		t.Fatalf("expected no error (pattern passes, leading digit ok for abc), got: %v", errs[0])
	}

	errs2 := fnVal.Validate("content", "output/blog/123abc.md", r)
	if len(errs2) == 0 {
		t.Fatal("expected error (pattern passes but leading digit blocked), got nil")
	}
	if !strings.Contains(errs2[0].Error(), "must not start with a digit") {
		t.Errorf("unexpected error: %v", errs2[0])
	}
}

func TestFileName_DefaultValidatorsIncludes(t *testing.T) {
	found := false
	for _, v := range validator.DefaultValidators() {
		if v.Name() == "file-name" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("DefaultValidators should include FileNameValidator")
	}
}
