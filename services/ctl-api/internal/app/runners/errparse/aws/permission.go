// Package aws holds the AWS provider-layer CompositeError parsers: cloud error
// codes surfaced from a runner job's raw output. They register at
// errparse.LayerProvider so a specific AWS cause (e.g. a missing IAM
// permission) wins over the tool-level fallback.
package aws

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

// AWSPermissionErrorType is the discriminator for AWS IAM permission failures
// surfaced from a terraform plan/apply.
const AWSPermissionErrorType compositeerrors.Type = "terraform.aws_permission"

// defaultIAMPolicyVersion is the IAM policy language version embedded in the
// remediation policy statement we recommend to users.
const defaultIAMPolicyVersion string = "2012-10-17"

// AWSPermissionError is the typed payload for an AWS API call that failed with
// AccessDenied / UnauthorizedOperation because the deploy's IAM principal is
// missing a permission.
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

// Hints tells the orchestrator not to auto-retry: a missing IAM permission
// won't resolve by retrying, so burning the auto-retry budget only delays
// surfacing the actionable error. The step is parked for manual retry once the
// user grants the permission.
func (e *AWSPermissionError) Hints() compositeerrors.Hints {
	return compositeerrors.NewHints().WithSkipAutoRetry()
}

// Sections returns the structured detail rendered in the dashboard: what AWS
// said, the principal/resource context, and a copy-pasteable IAM policy
// statement granting the missing action.
func (e *AWSPermissionError) Sections() []compositeerrors.Section {
	sections := []compositeerrors.Section{
		compositeerrors.MarkdownSection("Why", "The IAM principal used by this deployment was denied a permission required to perform the operation. This usually means the permission is not granted, but it can also be an explicit deny, a service control policy (SCP), or a permissions boundary. Grant or unblock the permission for the principal and retry."),
	}

	if e.RawMessage != "" {
		sections = append(sections, compositeerrors.CodeSection("AWS response", e.RawMessage))
	}

	if e.Principal != "" || e.Resource != "" {
		var lines []string
		if e.Principal != "" {
			lines = append(lines, fmt.Sprintf("Principal: %s", e.Principal))
		}
		if e.Resource != "" {
			lines = append(lines, fmt.Sprintf("Resource: %s", e.Resource))
		}
		sections = append(sections, compositeerrors.TextSection("Context", strings.Join(lines, "\n")))
	}

	if e.Action != "" {
		sections = append(sections,
			compositeerrors.MarkdownSection("How to fix", "Add the following statement to the role used by this deployment:"),
			compositeerrors.CodeSection("IAM policy statement", e.policyStatementJSON()),
		)
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

// awsPermissionPatterns are attempted in order. Each must define a named
// "action" group, and may define "principal" / "resource" / "code" groups.
var awsPermissionPatterns = []*regexp.Regexp{
	// Classic AccessDenied with principal + action + resource, e.g.
	// "AccessDenied: User: arn:aws:iam::123:role/nuon-runner is not authorized
	//  to perform: s3:CreateBucket on resource: arn:aws:s3:::acme-prod-assets"
	regexp.MustCompile(
		`(?P<code>AccessDenied(?:Exception)?|AuthorizationError):\s*(?:User|Principal):\s*(?P<principal>arn:[^\s]+)\s+is not authorized to perform:\s*(?P<action>[a-zA-Z0-9-]+:[a-zA-Z0-9*]+)(?:\s+on\s+resource:\s*(?P<resource>\S+))?`,
	),
	// Same shape without an explicit error-code prefix (some SDK clients).
	regexp.MustCompile(
		`(?:User|Principal):\s*(?P<principal>arn:[^\s]+)\s+is not authorized to perform:\s*(?P<action>[a-zA-Z0-9-]+:[a-zA-Z0-9*]+)(?:\s+on\s+resource:\s*(?P<resource>\S+))?`,
	),
	// EC2-style UnauthorizedOperation, where the action lives in a separate
	// sentence: "UnauthorizedOperation: ... Operation: ec2:CreateVpc"
	regexp.MustCompile(
		`(?P<code>UnauthorizedOperation):[^\n]*?(?:Operation|operation):\s*(?P<action>[a-zA-Z0-9-]+:[a-zA-Z0-9*]+)`,
	),
}

// permissionParser recognises AWS IAM permission failures in a runner job's
// raw output.
type permissionParser struct{}

func (permissionParser) Layer() errparse.Layer { return errparse.LayerProvider }

// Tools is nil (tool-agnostic): an AWS IAM denial is a provider-level failure
// that surfaces across tools, such as terraform and pulumi provisioning or an
// ECR push from a docker/oci build, so it must not be bucketed to a single
// tool. The AWS signals plus the provider check in Applicable are the real
// gates.
func (permissionParser) Tools() []errparse.Tool { return nil }
func (permissionParser) Signals() []string {
	return []string{
		"AccessDenied",
		"UnauthorizedOperation",
		"not authorized to perform",
		"AuthorizationError",
	}
}

// Applicable gates on the cloud provider, failing open on unknown.
func (permissionParser) Applicable(ctx *errparse.ParseContext) bool {
	switch ctx.Provider() {
	case errparse.ProviderAWS, errparse.ProviderUnknown:
		return true
	default:
		return false
	}
}

func (permissionParser) Parse(ctx *errparse.ParseContext) compositeerrors.CompositeError {
	raw := ctx.Raw
	for _, re := range awsPermissionPatterns {
		match := re.FindStringSubmatch(raw)
		if match == nil {
			continue
		}
		fields := groupMap(re, match)
		action := fields["action"]
		if action == "" {
			continue
		}

		return &AWSPermissionError{
			Action:       action,
			Resource:     cleanResource(fields["resource"]),
			Principal:    fields["principal"],
			AWSErrorCode: fields["code"],
			RawMessage:   extractRelevantLine(raw, match[0]),
		}
	}
	return nil
}

func init() {
	errparse.Register(permissionParser{})
}

func groupMap(re *regexp.Regexp, match []string) map[string]string {
	out := map[string]string{}
	for i, name := range re.SubexpNames() {
		if name == "" {
			continue
		}
		if i < len(match) {
			out[name] = match[i]
		}
	}
	return out
}

// cleanResource normalises a captured resource ARN. AWS quotes the ARN in some
// messages ("... on resource: \"arn:aws:s3:::bucket\" with an explicit deny
// ...") and glues sentence punctuation to it in others; both would otherwise
// leak into the "Context" section and the copy-pasteable IAM policy statement.
func cleanResource(s string) string {
	s = trimTrailingPunct(s)
	s = strings.Trim(s, `"'`)
	return trimTrailingPunct(s)
}

// trimTrailingPunct strips trailing punctuation that often glues to ARNs when
// AWS embeds them in sentences.
func trimTrailingPunct(s string) string {
	return strings.TrimRight(s, ".,;:")
}

// extractRelevantLine returns the line containing the match, trimmed of the
// terraform "│ " box-drawing prefix when present.
func extractRelevantLine(raw, matchStr string) string {
	needle := firstNChars(matchStr, 50)
	for _, line := range strings.Split(raw, "\n") {
		if strings.Contains(line, needle) {
			t := strings.TrimSpace(line)
			t = strings.TrimPrefix(t, "│")
			return strings.TrimSpace(t)
		}
	}
	return strings.TrimSpace(matchStr)
}

func firstNChars(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
