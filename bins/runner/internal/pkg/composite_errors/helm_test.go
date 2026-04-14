package composite_errors

import (
	"testing"
)

func TestParseHelmStderr_TemplateError(t *testing.T) {
	stderr := `Error: template: mychart/templates/deployment.yaml:12:5: executing "mychart/templates/deployment.yaml" at <.Values.missing>: nil pointer evaluating interface {}.missing`

	errors := ParseHelmStderr(stderr, "helm-install")

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	ce := errors[0]
	if ce.OwnerType != "helm-install" {
		t.Errorf("expected owner_type helm-install, got %s", ce.OwnerType)
	}
	if ce.Severity != "critical" {
		t.Errorf("expected severity critical, got %s", ce.Severity)
	}
	if ce.Metadata["template"] != "mychart/templates/deployment.yaml" {
		t.Errorf("expected template metadata, got %v", ce.Metadata["template"])
	}
	if ce.Metadata["line"] != "12" {
		t.Errorf("expected line 12, got %v", ce.Metadata["line"])
	}
	if ce.Metadata["column"] != "5" {
		t.Errorf("expected column 5, got %v", ce.Metadata["column"])
	}
}

func TestParseHelmStderr_InstallFailed(t *testing.T) {
	stderr := `Error: INSTALLATION FAILED: timed out waiting for the condition`

	errors := ParseHelmStderr(stderr, "helm-install")

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	ce := errors[0]
	if ce.Summary != "timed out waiting for the condition" {
		t.Errorf("unexpected summary: %s", ce.Summary)
	}
	if ce.Severity != "critical" {
		t.Errorf("expected severity critical, got %s", ce.Severity)
	}
}

func TestParseHelmStderr_UpgradeFailed(t *testing.T) {
	stderr := `Error: UPGRADE FAILED: release my-release does not exist`

	errors := ParseHelmStderr(stderr, "helm-upgrade")

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	if errors[0].Summary != "release my-release does not exist" {
		t.Errorf("unexpected summary: %s", errors[0].Summary)
	}
}

func TestParseHelmStderr_GenericError(t *testing.T) {
	stderr := `Error: could not find chart mychart in repo https://charts.example.com`

	errors := ParseHelmStderr(stderr, "helm-install")

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	if errors[0].Summary != "could not find chart mychart in repo https://charts.example.com" {
		t.Errorf("unexpected summary: %s", errors[0].Summary)
	}
}

func TestParseHelmStderr_MultipleErrors(t *testing.T) {
	stderr := `Error: INSTALLATION FAILED: first problem
Error: INSTALLATION FAILED: second problem`

	errors := ParseHelmStderr(stderr, "helm-install")

	if len(errors) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(errors))
	}
}

func TestParseHelmStderr_UnrecognizedPattern(t *testing.T) {
	stderr := `coalesce.go:286: warning: cannot overwrite table with non table for mykey
some other output
the final meaningful line`

	errors := ParseHelmStderr(stderr, "helm-install")

	if len(errors) != 1 {
		t.Fatalf("expected 1 fallback error, got %d", len(errors))
	}

	ce := errors[0]
	if ce.Summary != "the final meaningful line" {
		t.Errorf("expected last non-empty line as summary, got %s", ce.Summary)
	}
	if ce.Detail != stderr {
		t.Errorf("expected full stderr as detail")
	}
}

func TestParseHelmStderr_EmptyInput(t *testing.T) {
	if errors := ParseHelmStderr("", "helm-install"); errors != nil {
		t.Errorf("expected nil for empty input, got %v", errors)
	}
	if errors := ParseHelmStderr("   \n  \n  ", "helm-install"); errors != nil {
		t.Errorf("expected nil for whitespace-only input, got %v", errors)
	}
}
