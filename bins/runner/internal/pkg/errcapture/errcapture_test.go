package errcapture

import (
	"context"
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
