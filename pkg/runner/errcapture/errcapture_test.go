package errcapture

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"unicode/utf8"

	"go.uber.org/zap"
)

func TestCore_CapturesErrorLevelOnly(t *testing.T) {
	c := New()
	l := zap.New(c.Core())

	l.Debug("debug line")
	l.Info("info line")
	l.Warn("warn line")
	l.Error("boom")

	got := c.String()
	if got != "boom" {
		t.Fatalf("captured %q, want only the error line", got)
	}
}

func TestCore_AppendsErrorFieldToMessage(t *testing.T) {
	c := New()
	l := zap.New(c.Core())

	l.Error("terraform run errored", zap.Error(errors.New("exit status 1")))

	if got := c.String(); got != "terraform run errored: exit status 1" {
		t.Fatalf("captured %q", got)
	}
}

func TestCore_ErrorFieldFromWith(t *testing.T) {
	c := New()
	l := zap.New(c.Core()).With(zap.Error(errors.New("deep cause")))

	l.Error("wrapped")

	if got := c.String(); got != "wrapped: deep cause" {
		t.Fatalf("captured %q, want the With() error surfaced", got)
	}
}

func TestCore_JoinsMultipleLines(t *testing.T) {
	c := New()
	l := zap.New(c.Core())

	l.Error("Error: creating S3 Bucket (x): AccessDenied")
	l.Error("Error: creating S3 Bucket (y): AccessDenied")

	got := c.String()
	if !strings.Contains(got, "(x)") || !strings.Contains(got, "(y)") {
		t.Fatalf("expected both lines, got %q", got)
	}
	if strings.Count(got, "\n") != 1 {
		t.Fatalf("expected 2 joined lines, got %q", got)
	}
}

// terraformDiagnosticLine is a synthetic terraform JSON log record of the shape
// emitted for a failed apply, with the real cause in diagnostic.detail.
const terraformDiagnosticLine = `{
  "@level": "error",
  "@message": "Error: creating S3 Bucket (acme-artifacts): AccessDenied",
  "@module": "terraform.ui",
  "@timestamp": "2026-01-01T00:00:00.000000Z",
  "diagnostic": {
    "severity": "error",
    "summary": "creating S3 Bucket (acme-artifacts): AccessDenied",
    "detail": "User: arn:aws:sts::000000000000:assumed-role/acme/runner is not authorized to perform: s3:CreateBucket on resource: arn:aws:s3:::acme-artifacts",
    "address": "module.storage.aws_s3_bucket.artifacts"
  },
  "type": "diagnostic"
}`

// logTerraformLine mirrors what pkg/zaphclog does with a terraform JSON record:
// the @-prefixed keys drive the level and message, every other key is logged as
// a zap.Any field.
func logTerraformLine(t *testing.T, l *zap.Logger, line string) {
	t.Helper()

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(line), &obj); err != nil {
		t.Fatalf("unable to parse terraform line: %v", err)
	}

	var fields []zap.Field
	for k, v := range obj {
		if strings.HasPrefix(k, "@") {
			continue
		}
		fields = append(fields, zap.Any(k, v))
	}

	msg, _ := obj["@message"].(string)
	l.Error(msg, fields...)
}

func TestCore_CapturesTerraformDiagnosticDetailAndAddress(t *testing.T) {
	c := New()
	logTerraformLine(t, zap.New(c.Core()), terraformDiagnosticLine)

	got := c.String()
	want := "Error: creating S3 Bucket (acme-artifacts): AccessDenied\n" +
		"  with module.storage.aws_s3_bucket.artifacts\n" +
		"User: arn:aws:sts::000000000000:assumed-role/acme/runner is not authorized to perform: s3:CreateBucket on resource: arn:aws:s3:::acme-artifacts"
	if got != want {
		t.Fatalf("captured:\n%s\n\nwant:\n%s", got, want)
	}
}

func TestCore_TerraformDiagnosticSummaryFallbackWhenMessageEmpty(t *testing.T) {
	c := New()
	l := zap.New(c.Core())

	l.Error("", zap.Any("diagnostic", map[string]interface{}{
		"severity": "error",
		"summary":  "Invalid function argument",
	}))

	if got := c.String(); got != "Error: Invalid function argument" {
		t.Fatalf("captured %q, want the summary used as the headline", got)
	}
}

func TestCore_TerraformDiagnosticSummaryNotDuplicated(t *testing.T) {
	c := New()
	l := zap.New(c.Core())

	l.Error("Error: Invalid function argument", zap.Any("diagnostic", map[string]interface{}{
		"summary": "Invalid function argument",
		"detail":  "Call to function \"file\" failed: no file exists at \"missing.txt\".",
	}))

	got := c.String()
	if strings.Count(got, "Invalid function argument") != 1 {
		t.Fatalf("captured %q, want the summary kept once", got)
	}
	if !strings.Contains(got, "no file exists") {
		t.Fatalf("captured %q, want the detail appended", got)
	}
}

func TestCore_IgnoresNonDiagnosticField(t *testing.T) {
	c := New()
	l := zap.New(c.Core())

	l.Error("boom", zap.Any("diagnostic", "not an object"))

	if got := c.String(); got != "boom" {
		t.Fatalf("captured %q, want the message unchanged", got)
	}
}

func TestCore_TerraformDiagnosticRespectsBound(t *testing.T) {
	c := &Capture{max: 32}
	logTerraformLine(t, zap.New(c.Core()), terraformDiagnosticLine)

	got := c.String()
	if len(got) > 32 {
		t.Fatalf("captured %d bytes, want at most 32", len(got))
	}
	if !utf8.ValidString(got) {
		t.Fatalf("captured %q is not valid UTF-8", got)
	}
	if !strings.HasPrefix(got, "Error: creating S3 Bucket") {
		t.Fatalf("captured %q, want the head of the diagnostic preserved", got)
	}
}

func TestAppend_BoundedKeepsHeadAndTruncatesOverflow(t *testing.T) {
	c := &Capture{max: 12} // "aaaa"(4)+1 then "bbbb"(4)+1 = 10; "cccc" overflows, 1 byte budget left
	c.append("aaaa")
	c.append("bbbb")
	c.append("cccc")

	got := c.String()
	if got != "aaaa\nbbbb\nc" {
		t.Fatalf("captured %q, want head preserved and overflow truncated to fit", got)
	}
	if !c.full {
		t.Fatal("expected capture to be marked full after overflow")
	}

	c.append("dddd")
	if got := c.String(); got != "aaaa\nbbbb\nc" {
		t.Fatalf("captured %q, want no further appends after full", got)
	}
}

func TestAppend_FirstLineOversizedKeepsTruncatedPrefix(t *testing.T) {
	c := &Capture{max: 5}
	c.append("Error: something very long that exceeds the bound")

	got := c.String()
	if got != "Erro" { // budget = 5 - 0 - 1 = 4
		t.Fatalf("captured %q, want a non-empty truncated prefix so the root cause survives", got)
	}
}

func TestAppend_OversizedMultibyteDoesNotSplitRune(t *testing.T) {
	// "€" is 3 bytes. With budget 4 we can only fit one full rune (3 bytes),
	// never a partial one.
	c := &Capture{max: 5}
	c.append("€€€")

	got := c.String()
	if got != "€" {
		t.Fatalf("captured %q, want a single full rune (no byte-split)", got)
	}
	if !utf8.ValidString(got) {
		t.Fatalf("captured %q is not valid UTF-8", got)
	}
}

func TestString_NilSafe(t *testing.T) {
	var c *Capture
	if c.String() != "" {
		t.Fatal("nil Capture should stringify to empty")
	}
}

func TestContext_RoundTrip(t *testing.T) {
	c := New()
	c.append("x")
	ctx := NewContext(context.Background(), c)

	if FromContext(ctx) != c {
		t.Fatal("FromContext did not return the stored Capture")
	}
	if Output(ctx) != "x" {
		t.Fatalf("Output = %q", Output(ctx))
	}
	if Output(context.Background()) != "" {
		t.Fatal("Output on a bare context should be empty")
	}
}
