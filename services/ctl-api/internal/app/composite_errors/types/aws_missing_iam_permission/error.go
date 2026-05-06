// Package aws_missing_iam_permission is the cross-cutting CompositeError for
// AWS API calls that fail with AccessDenied / UnauthorizedOperation errors
// because the caller's IAM principal is missing a permission.
//
// This is the showcase parser: it can be triggered from terraform-plan,
// terraform-apply, helm-install (when EKS providers fail to assume a role),
// or runner-job output (when the runner itself lacks a permission). The
// parser registers against multiple ParseContext subtrees rather than a
// single hierarchy node.
package aws_missing_iam_permission

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	composite_error "github.com/nuonco/nuon/pkg/composite_error"
)

const Type composite_error.Type = "aws_missing_iam_permission"

// Error is the typed payload.
type Error struct {
	// Action is the IAM action the caller lacked, e.g. "ec2:CreateVpc",
	// "s3:GetObject". May be empty if extraction failed (rare; the parser
	// only matches when it's confident enough to populate this).
	Action string `json:"action"`

	// Resource is the ARN (or wildcard) the call targeted, when known.
	Resource string `json:"resource,omitempty"`

	// Principal is the IAM principal ARN the call was made as, when known.
	Principal string `json:"principal,omitempty"`

	// AWSErrorCode is the API error code we matched on (AccessDenied,
	// UnauthorizedOperation, AccessDeniedException, …). Useful for the UI
	// when surfacing related runbook links.
	AWSErrorCode string `json:"aws_error_code,omitempty"`

	// RawMessage is the AWS-emitted error message we extracted the fields from.
	RawMessage string `json:"raw_message,omitempty"`
}

var _ composite_error.CompositeError = (*Error)(nil)
var _ composite_error.ErrorWithDirective = (*Error)(nil)
var _ composite_error.ErrorWithDocsLink = (*Error)(nil)

func (e *Error) Type() composite_error.Type         { return Type }
func (e *Error) Domain() composite_error.Domain     { return composite_error.DomainAWS }
func (e *Error) Severity() composite_error.Severity { return composite_error.SeverityError }

// OverrideDirective tells the conductor to stop retrying — auto-retrying an
// IAM permission failure without user action is wasted work.
func (e *Error) OverrideDirective() composite_error.Directive {
	return composite_error.Directive{Kind: composite_error.DirectiveStop}
}

func (e *Error) DocsURL() string {
	return "https://docs.nuon.co/troubleshooting/aws-iam-permissions"
}

func (e *Error) Render(_ context.Context) composite_error.Render {
	r := composite_error.Render{}

	if e.Action != "" {
		r.Title = fmt.Sprintf("Missing AWS IAM permission: %s", e.Action)
	} else {
		r.Title = "Missing AWS IAM permission"
	}

	r.Summary = "The IAM principal used by this deployment is missing a permission required to perform the operation. Grant the permission to the principal and retry."

	if e.RawMessage != "" {
		r.Sections = append(r.Sections, composite_error.RenderSection{
			Heading: "AWS response",
			Body:    "```\n" + e.RawMessage + "\n```",
		})
	}

	// Show the principal/resource so the user knows what role to edit.
	if e.Principal != "" || e.Resource != "" {
		var b strings.Builder
		if e.Principal != "" {
			fmt.Fprintf(&b, "Principal: `%s`\n", e.Principal)
		}
		if e.Resource != "" {
			fmt.Fprintf(&b, "Resource: `%s`\n", e.Resource)
		}
		r.Sections = append(r.Sections, composite_error.RenderSection{
			Heading: "Context",
			Body:    strings.TrimRight(b.String(), "\n"),
		})
	}

	// Always offer a copy-pasteable IAM policy fragment when we have the action.
	if e.Action != "" {
		r.Sections = append(r.Sections, composite_error.RenderSection{
			Heading: "Suggested IAM policy statement",
			Body:    "Add the following to the role used by this deployment:\n\n```json\n" + e.policyStatementJSON() + "\n```",
		})
		r.UserActions = append(r.UserActions, composite_error.UserAction{
			Kind:  composite_error.UserActionKindCopy,
			Label: "Copy IAM policy statement",
			Value: e.policyStatementJSON(),
		})
	}

	r.UserActions = append(r.UserActions,
		composite_error.UserAction{
			Kind:  composite_error.UserActionKindLink,
			Label: "Troubleshooting docs",
			Value: e.DocsURL(),
		},
		composite_error.UserAction{
			Kind:  composite_error.UserActionKindRetry,
			Label: "Retry after granting the permission",
		},
	)

	return r
}

// policyStatementJSON renders a minimal IAM policy statement granting the
// missing action on the resource (or "*" if the resource isn't known).
func (e *Error) policyStatementJSON() string {
	resource := e.Resource
	if resource == "" {
		resource = "*"
	}
	stmt := map[string]any{
		"Version": "2012-10-17",
		"Statement": []map[string]any{
			{
				"Effect":   "Allow",
				"Action":   []string{e.Action},
				"Resource": resource,
			},
		},
	}
	b, _ := json.MarshalIndent(stmt, "", "  ")
	return string(b)
}
