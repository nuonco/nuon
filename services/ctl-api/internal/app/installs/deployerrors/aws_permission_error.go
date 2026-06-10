// Package deployerrors holds the typed CompositeError implementations produced
// by component deploy (plan + apply) flows. It lives next to the installs
// domain that consumes it rather than in a central error catalog: each domain
// owns the custom errors it knows how to produce and parse.
package deployerrors

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

// AWSPermissionErrorType is the discriminator for AWS IAM permission failures
// surfaced from a terraform plan/apply during a component deploy.
const AWSPermissionErrorType compositeerrors.Type = "terraform.aws_permission"

// defaultIAMPolicyVersion is the IAM policy language version embedded in the
// remediation policy statement we recommend to users.
const defaultIAMPolicyVersion string = "2012-10-17"

// AWSPermissionError is the typed payload for an AWS API call that failed with
// AccessDenied / UnauthorizedOperation because the deploy's IAM principal is
// missing a permission. It implements compositeerrors.CompositeError so it can
// be returned like any error and frozen onto the owning deploy row.
type AWSPermissionError struct {
	// Action is the IAM action the caller lacked, e.g. "ec2:CreateVpc",
	// "s3:CreateBucket".
	Action string `json:"action"`

	// Resource is the ARN (or wildcard) the call targeted, when known.
	Resource string `json:"resource,omitempty"`

	// Principal is the IAM principal ARN the call was made as, when known.
	Principal string `json:"principal,omitempty"`

	// AWSErrorCode is the API error code we matched on (AccessDenied,
	// UnauthorizedOperation, AccessDeniedException, AuthorizationError).
	AWSErrorCode string `json:"aws_error_code,omitempty"`

	// RawMessage is the AWS-emitted error line we extracted the fields from.
	RawMessage string `json:"raw_message,omitempty"`
}

var _ compositeerrors.CompositeError = (*AWSPermissionError)(nil)

// Error returns the one-line headline shown to users.
func (e *AWSPermissionError) Error() string {
	if e.Action != "" {
		return fmt.Sprintf("Missing AWS IAM permission: %s", e.Action)
	}
	return "Missing AWS IAM permission"
}

func (e *AWSPermissionError) Type() compositeerrors.Type { return AWSPermissionErrorType }
func (e *AWSPermissionError) Severity() compositeerrors.Severity {
	return compositeerrors.SeverityError
}

// Sections returns the structured detail rendered in the dashboard: what AWS
// said, the principal/resource context, and a copy-pasteable IAM policy
// statement granting the missing action.
func (e *AWSPermissionError) Sections() []compositeerrors.Section {
	sections := []compositeerrors.Section{
		{
			Heading: "Why",
			Body:    "The IAM principal used by this deployment was denied a permission required to perform the operation. This usually means the permission is not granted, but it can also be an explicit deny, a service control policy (SCP), or a permissions boundary. Grant or unblock the permission for the principal and retry.",
		},
	}

	if e.RawMessage != "" {
		sections = append(sections, compositeerrors.Section{
			Heading: "AWS response",
			Body:    "```\n" + e.RawMessage + "\n```",
		})
	}

	if e.Principal != "" || e.Resource != "" {
		var lines []string
		if e.Principal != "" {
			lines = append(lines, fmt.Sprintf("Principal: `%s`", e.Principal))
		}
		if e.Resource != "" {
			lines = append(lines, fmt.Sprintf("Resource: `%s`", e.Resource))
		}
		sections = append(sections, compositeerrors.Section{
			Heading: "Context",
			Body:    strings.Join(lines, "\n\n"),
		})
	}

	if e.Action != "" {
		sections = append(sections, compositeerrors.Section{
			Heading: "How to fix",
			Body:    "Add the following to the role used by this deployment:\n\n```json\n" + e.policyStatementJSON() + "\n```",
		})
	}

	return sections
}

// policyStatementJSON renders a minimal IAM policy statement granting the
// missing action on the resource (or "*" if the resource isn't known).
func (e *AWSPermissionError) policyStatementJSON() string {
	resource := e.Resource
	if resource == "" {
		resource = "*"
	}
	stmt := map[string]any{
		"Version": defaultIAMPolicyVersion,
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
