package composite_errors

import (
	"testing"
)

func TestParseTerraformJSON_DiagnosticError(t *testing.T) {
	input := []byte(`{"@level":"error","@message":"Error: Invalid reference","@module":"terraform.ui","type":"diagnostic","diagnostic":{"severity":"error","summary":"Invalid reference","detail":"A reference to var.missing was not found.","address":"aws_instance.web","range":{"filename":"main.tf","start":{"line":42,"column":5},"end":{"line":42,"column":20}}}}`)

	errors := ParseTerraformJSON(input, "plan")

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	ce := errors[0]
	if ce.OwnerType != "plan" {
		t.Errorf("expected owner_type terraform-plan, got %s", ce.OwnerType)
	}
	if ce.Severity != "critical" {
		t.Errorf("expected severity critical, got %s", ce.Severity)
	}
	if ce.Summary != "Invalid reference" {
		t.Errorf("expected summary 'Invalid reference', got %s", ce.Summary)
	}
	if ce.Detail != "A reference to var.missing was not found." {
		t.Errorf("unexpected detail: %s", ce.Detail)
	}
	if ce.Metadata["resource"] != "aws_instance.web" {
		t.Errorf("expected resource aws_instance.web, got %v", ce.Metadata["resource"])
	}
	if ce.Metadata["file"] != "main.tf" {
		t.Errorf("expected file main.tf, got %v", ce.Metadata["file"])
	}
	if ce.Metadata["line"] != 42 {
		t.Errorf("expected line 42, got %v", ce.Metadata["line"])
	}
	if ce.Metadata["column"] != 5 {
		t.Errorf("expected column 5, got %v", ce.Metadata["column"])
	}
}

func TestParseTerraformJSON_Warning(t *testing.T) {
	input := []byte(`{"@level":"warn","@message":"Warning: Deprecated attribute","@module":"terraform.ui","type":"diagnostic","diagnostic":{"severity":"warning","summary":"Deprecated attribute","detail":"Use new_attr instead."}}`)

	errors := ParseTerraformJSON(input, "apply")

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	ce := errors[0]
	if ce.Severity != "warning" {
		t.Errorf("expected severity warning, got %s", ce.Severity)
	}
	if ce.Summary != "Deprecated attribute" {
		t.Errorf("expected summary 'Deprecated attribute', got %s", ce.Summary)
	}
	if ce.Metadata != nil {
		t.Errorf("expected nil metadata for warning without address/range, got %v", ce.Metadata)
	}
}

func TestParseTerraformJSON_MixedOutput(t *testing.T) {
	input := []byte(`{"@level":"info","@message":"Terraform has been successfully initialized!","@module":"terraform.ui","type":"version"}
{"@level":"info","@message":"Plan: 1 to add","@module":"terraform.ui","type":"change_summary"}
{"@level":"error","@message":"Error: Missing variable","@module":"terraform.ui","type":"diagnostic","diagnostic":{"severity":"error","summary":"Missing variable","detail":"Variable 'region' is required."}}`)

	errors := ParseTerraformJSON(input, "plan")

	if len(errors) != 1 {
		t.Fatalf("expected 1 error from mixed output, got %d", len(errors))
	}

	if errors[0].Summary != "Missing variable" {
		t.Errorf("expected summary 'Missing variable', got %s", errors[0].Summary)
	}
}

func TestParseTerraformJSON_MultipleErrors(t *testing.T) {
	input := []byte(`{"@level":"error","@message":"Error: Invalid reference","@module":"terraform.ui","type":"diagnostic","diagnostic":{"severity":"error","summary":"Invalid reference","detail":"bad ref"}}
{"@level":"warn","@message":"Warning: Deprecated","@module":"terraform.ui","type":"diagnostic","diagnostic":{"severity":"warning","summary":"Deprecated feature","detail":"use new thing"}}`)

	errors := ParseTerraformJSON(input, "plan")

	if len(errors) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(errors))
	}
	if errors[0].Severity != "critical" {
		t.Errorf("first error should be critical, got %s", errors[0].Severity)
	}
	if errors[1].Severity != "warning" {
		t.Errorf("second error should be warning, got %s", errors[1].Severity)
	}
}

func TestParseTerraformJSON_EmptyInput(t *testing.T) {
	if errors := ParseTerraformJSON(nil, "plan"); errors != nil {
		t.Errorf("expected nil for nil input, got %v", errors)
	}
	if errors := ParseTerraformJSON([]byte{}, "plan"); errors != nil {
		t.Errorf("expected nil for empty input, got %v", errors)
	}
}

func TestParseTerraformJSON_MalformedJSON(t *testing.T) {
	input := []byte(`not json at all
{"@level":"error","@message":"Error: Bad thing","@module":"terraform.ui","type":"diagnostic","diagnostic":{"severity":"error","summary":"Bad thing","detail":"details"}}
{truncated json...`)

	errors := ParseTerraformJSON(input, "plan")

	if len(errors) != 1 {
		t.Fatalf("expected 1 error (malformed lines skipped), got %d", len(errors))
	}
	if errors[0].Summary != "Bad thing" {
		t.Errorf("expected summary 'Bad thing', got %s", errors[0].Summary)
	}
}

func TestParseTerraformJSON_ErrorLevelWithoutDiagnostic(t *testing.T) {
	input := []byte(`{"@level":"error","@message":"Error: Failed to load backend","@module":"terraform.ui","type":"error_message"}`)

	errors := ParseTerraformJSON(input, "terraform-init")

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
	if errors[0].Summary != "Error: Failed to load backend" {
		t.Errorf("expected message as summary, got %s", errors[0].Summary)
	}
	if errors[0].Severity != "critical" {
		t.Errorf("expected critical severity, got %s", errors[0].Severity)
	}
}
