package terraform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func readFixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("unable to read fixture %s: %v", name, err)
	}
	return string(b)
}

func parse(raw string) compositeerrors.CompositeError {
	return errorParser{}.Parse(&errparse.ParseContext{Raw: raw})
}

func TestParse_SingleError(t *testing.T) {
	ce := parse(readFixture(t, "invalid_reference.txt"))
	if ce == nil {
		t.Fatal("expected a composite error, got nil")
	}
	e, ok := ce.(*TerraformError)
	if !ok {
		t.Fatalf("expected *TerraformError, got %T", ce)
	}
	if e.Summary != "Reference to undeclared resource" {
		t.Errorf("summary = %q", e.Summary)
	}
	if len(e.Errors) != 0 {
		t.Errorf("expected no Errors list for a single error, got %v", e.Errors)
	}
	if ce.Error() != "Reference to undeclared resource" {
		t.Errorf("headline = %q", ce.Error())
	}
	if ce.Type() != TerraformErrorType {
		t.Errorf("type = %q", ce.Type())
	}

	headings := map[string]string{}
	for _, s := range ce.Sections() {
		headings[s.Heading] = s.Body
	}
	if _, ok := headings["Errors"]; ok {
		t.Error("single error should not render an Errors list section")
	}
	out, ok := headings["Output"]
	if !ok {
		t.Fatal("expected an Output section")
	}
	// The output preserves the detail (with the "│" box-drawing stripped) and
	// drops our own wrapper is fine to retain — but the terraform detail must
	// survive.
	if !strings.Contains(out, `A managed resource "aws_subnet" "main" has not been declared`) {
		t.Errorf("Output missing terraform detail: %q", out)
	}
	if strings.Contains(out, "│") {
		t.Errorf("Output still carries box-drawing: %q", out)
	}
}

func TestParse_MultipleErrors(t *testing.T) {
	ce := parse(readFixture(t, "multiple_bucket_errors.txt"))
	if ce == nil {
		t.Fatal("expected a composite error, got nil")
	}
	e := ce.(*TerraformError)
	if len(e.Errors) != 2 {
		t.Fatalf("expected 2 distinct errors, got %d: %v", len(e.Errors), e.Errors)
	}
	if !strings.HasPrefix(e.Summary, "creating S3 Bucket (acme-logs)") {
		t.Errorf("summary = %q", e.Summary)
	}
	if !strings.Contains(ce.Error(), "(+1 more errors)") {
		t.Errorf("headline should note the extra error: %q", ce.Error())
	}

	var errorsSection string
	for _, s := range ce.Sections() {
		if s.Heading == "Errors" {
			errorsSection = s.Body
		}
	}
	if errorsSection == "" {
		t.Fatal("expected an Errors list section")
	}
	if !strings.Contains(errorsSection, "acme-logs") || !strings.Contains(errorsSection, "acme-data") {
		t.Errorf("Errors section missing summaries: %q", errorsSection)
	}
}

func TestParse_NoTerraformError(t *testing.T) {
	cases := []string{
		"",
		"terraform run errored",
		"job step errored unable to execute job: exit status 1",
		"error running apply: exit status 1",
	}
	for _, in := range cases {
		if ce := parse(in); ce != nil {
			t.Errorf("parse(%q) = %v, want nil (no Error: diagnostic)", in, ce)
		}
	}
}
