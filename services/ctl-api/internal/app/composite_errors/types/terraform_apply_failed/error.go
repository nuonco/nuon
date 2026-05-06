// Package terraform_apply_failed is the broad CompositeError type for a failed
// terraform run (init / plan / apply / destroy).
//
// It is the "generic" fallback for the terraform domain — when a more specific
// parser (e.g. aws_missing_iam_permission) doesn't claim the input, this one
// catches it and surfaces the resource-level diagnostics terraform prints.
package terraform_apply_failed

import (
	"context"
	"fmt"
	"strings"

	composite_error "github.com/nuonco/nuon/pkg/composite_error"
)

const Type composite_error.Type = "terraform_apply_failed"

// Diagnostic is one resource-level error block emitted by terraform.
//
// Terraform's CLI output for an error block looks like:
//
//	╷
//	│ Error: creating EC2 Subnet: ...
//	│
//	│   with module.vpc.aws_subnet.public[0],
//	│   on .terraform/modules/vpc/main.tf line 218, in resource "aws_subnet" "public":
//	│  218: resource "aws_subnet" "public" {
//	│
//	╵
//
// We capture the structured fields so the dashboard can render them as a
// table; the raw block is preserved for fidelity.
type Diagnostic struct {
	Summary    string `json:"summary"`               // first "Error:" line, after the prefix
	Resource   string `json:"resource,omitempty"`    // e.g. "module.vpc.aws_subnet.public[0]"
	SourceFile string `json:"source_file,omitempty"` // e.g. ".terraform/modules/vpc/main.tf"
	SourceLine int    `json:"source_line,omitempty"`
	Raw        string `json:"raw,omitempty"`
}

// Error is the typed payload for a terraform_apply_failed CompositeError.
type Error struct {
	// Stage is "init" | "plan" | "apply" | "destroy" — best effort from
	// the parser context.
	Stage string `json:"stage,omitempty"`

	// Diagnostics extracted from the terraform output. May be empty if no
	// structured diagnostics were found, in which case Message holds the
	// first line of the raw output.
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`

	// Message is a fallback short message when no diagnostics were extracted.
	Message string `json:"message,omitempty"`
}

var _ composite_error.CompositeError = (*Error)(nil)

func (e *Error) Type() composite_error.Type         { return Type }
func (e *Error) Domain() composite_error.Domain     { return composite_error.DomainTerraform }
func (e *Error) Severity() composite_error.Severity { return composite_error.SeverityError }

func (e *Error) Render(_ context.Context) composite_error.Render {
	stage := e.Stage
	if stage == "" {
		stage = "apply"
	}

	r := composite_error.Render{}

	switch len(e.Diagnostics) {
	case 0:
		r.Title = fmt.Sprintf("Terraform %s failed", stage)
		if e.Message != "" {
			r.Summary = e.Message
		}
	case 1:
		r.Title = fmt.Sprintf("Terraform %s failed: %s", stage, truncate(e.Diagnostics[0].Summary, 100))
	default:
		r.Title = fmt.Sprintf("Terraform %s failed with %d errors", stage, len(e.Diagnostics))
		summaries := make([]string, 0, len(e.Diagnostics))
		for _, d := range e.Diagnostics {
			summaries = append(summaries, "• "+d.Summary)
		}
		r.Summary = strings.Join(summaries, "\n")
	}

	for _, d := range e.Diagnostics {
		body := d.Summary
		if d.Resource != "" {
			body += fmt.Sprintf("\n\nResource: `%s`", d.Resource)
		}
		if d.SourceFile != "" {
			body += fmt.Sprintf("\nLocation: `%s`", d.SourceFile)
			if d.SourceLine > 0 {
				body += fmt.Sprintf(" (line %d)", d.SourceLine)
			}
		}
		r.Sections = append(r.Sections, composite_error.RenderSection{
			Heading: "Error",
			Body:    body,
		})
	}

	return r
}

func truncate(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
