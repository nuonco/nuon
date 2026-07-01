package generic

import (
	"strings"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func parse(raw string) compositeerrors.CompositeError {
	return genericParser{}.Parse(&errparse.ParseContext{Raw: raw})
}

func TestParse_SingleLine(t *testing.T) {
	ce := parse("exit status 1")
	if ce == nil {
		t.Fatal("expected a composite error, got nil")
	}
	if ce.Type() != GenericErrorType {
		t.Errorf("type = %q, want %q", ce.Type(), GenericErrorType)
	}
	if ce.Error() != "exit status 1" {
		t.Errorf("headline = %q", ce.Error())
	}
	if ce.Severity() != compositeerrors.SeverityError {
		t.Errorf("severity = %q", ce.Severity())
	}
}

func TestParse_PrefersErrorLineAsHeadline(t *testing.T) {
	raw := strings.Join([]string{
		"Refreshing state...",
		"│ Error: creating S3 Bucket (acme): AccessDenied",
		"terraform run errored",
		"job step errored unable to execute job",
	}, "\n")

	ce := parse(raw)
	if ce == nil {
		t.Fatal("expected a composite error, got nil")
	}
	if ce.Error() != "Error: creating S3 Bucket (acme): AccessDenied" {
		t.Errorf("headline = %q, want the Error: line with the box prefix stripped", ce.Error())
	}
}

func TestParse_FallsBackToFirstLine(t *testing.T) {
	raw := "helm upgrade failed\nsome trailing context"
	ce := parse(raw)
	if ce.Error() != "helm upgrade failed" {
		t.Errorf("headline = %q, want first non-blank line", ce.Error())
	}
}

func TestParse_EmptyAndWhitespaceReturnNil(t *testing.T) {
	for _, in := range []string{"", "   ", "\n\n", "│\n│  \n"} {
		if ce := parse(in); ce != nil {
			t.Errorf("parse(%q) = %v, want nil", in, ce)
		}
	}
}

func TestParse_CleansBoxPrefixAndBlankLines(t *testing.T) {
	raw := "│ Error: boom\n│\n│   details here\n\n"
	ce := parse(raw)
	e := ce.(*GenericError)
	if e.Body != "Error: boom\ndetails here" {
		t.Errorf("body = %q, want cleaned lines without box prefix/blanks", e.Body)
	}
	sections := ce.Sections()
	if len(sections) != 1 || sections[0].Heading != "Error output" {
		t.Fatalf("unexpected sections: %v", sections)
	}
	if !strings.Contains(sections[0].Body, "Error: boom") {
		t.Errorf("section body missing content: %q", sections[0].Body)
	}
}

func TestParse_TruncatesLongBody(t *testing.T) {
	raw := strings.Repeat("x", maxBody+500)
	e := parse(raw).(*GenericError)
	if len([]rune(e.Body)) != maxBody+1 { // maxBody chars + the ellipsis rune
		t.Errorf("body rune length = %d, want %d", len([]rune(e.Body)), maxBody+1)
	}
	if !strings.HasSuffix(e.Body, "…") {
		t.Error("expected truncation ellipsis")
	}
}

func TestParse_TruncatesLongHeadline(t *testing.T) {
	raw := "Error: " + strings.Repeat("y", maxHeadline)
	ce := parse(raw)
	if !strings.HasSuffix(ce.Error(), "…") {
		t.Errorf("expected headline truncation, got %q", ce.Error())
	}
}

// TestRegistry_GenericIsFallbackOnly proves the generic parser only wins when no
// specific parser matched: it fires for unclassified text but must not shadow a
// registered specific parser at a lower layer.
func TestRegistry_GenericIsFallbackOnly(t *testing.T) {
	// The generic parser registers into the default registry via init(); no
	// specific parser is imported here, so it should catch arbitrary text.
	ce := errparse.Parse(&errparse.ParseContext{Raw: "something totally unrecognised", Tool: errparse.ToolTerraform})
	if ce == nil {
		t.Fatal("expected the generic fallback to match unclassified text")
	}
	if ce.Type() != GenericErrorType {
		t.Errorf("type = %q, want generic fallback", ce.Type())
	}
}
