package composite_errors

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseActionOutput_BasicOutput(t *testing.T) {
	output := `running migration step 1...
running migration step 2...
FATAL: connection refused`

	errors := ParseActionOutput(output, "action-run")

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	ce := errors[0]
	if ce.OwnerType != "action-run" {
		t.Errorf("expected owner_type action-run, got %s", ce.OwnerType)
	}
	if ce.Severity != "critical" {
		t.Errorf("expected severity critical, got %s", ce.Severity)
	}
	if ce.Summary != "FATAL: connection refused" {
		t.Errorf("expected last line as summary, got %s", ce.Summary)
	}
	if ce.Detail != output {
		t.Errorf("expected full output as detail since <= 20 lines")
	}
}

func TestParseActionOutput_TruncatesDetail(t *testing.T) {
	var lines []string
	for i := 1; i <= 30; i++ {
		lines = append(lines, fmt.Sprintf("line %d", i))
	}
	output := strings.Join(lines, "\n")

	errors := ParseActionOutput(output, "action-run")

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	ce := errors[0]
	if ce.Summary != "line 30" {
		t.Errorf("expected last line as summary, got %s", ce.Summary)
	}

	detailLines := strings.Split(ce.Detail, "\n")
	if len(detailLines) != 20 {
		t.Errorf("expected 20 lines in detail, got %d", len(detailLines))
	}
	if detailLines[0] != "line 11" {
		t.Errorf("expected detail to start at line 11, got %s", detailLines[0])
	}
}

func TestParseActionOutput_SingleLine(t *testing.T) {
	output := "command not found: deploy"

	errors := ParseActionOutput(output, "action-run")

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
	if errors[0].Summary != "command not found: deploy" {
		t.Errorf("unexpected summary: %s", errors[0].Summary)
	}
}

func TestParseActionOutput_EmptyInput(t *testing.T) {
	if errors := ParseActionOutput("", "action-run"); errors != nil {
		t.Errorf("expected nil for empty input, got %v", errors)
	}
	if errors := ParseActionOutput("   \n  \n  ", "action-run"); errors != nil {
		t.Errorf("expected nil for whitespace-only input, got %v", errors)
	}
}

func TestParseActionOutput_TrailingNewlines(t *testing.T) {
	output := "some output\nthe error\n\n\n"

	errors := ParseActionOutput(output, "action-run")

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
	if errors[0].Summary != "the error" {
		t.Errorf("expected 'the error' as summary (skipping trailing empty lines), got %s", errors[0].Summary)
	}
}
