package aws

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

const AWSPermissionErrorType compositeerrors.Type = "terraform.aws_permission"

const defaultIAMPolicyVersion string = "2012-10-17"

type AWSPermissionError struct {
	Action string `json:"action"`

	Resource string `json:"resource,omitempty"`

	Principal string `json:"principal,omitempty"`

	AWSErrorCode string `json:"aws_error_code,omitempty"`

	RawMessage string `json:"raw_message,omitempty"`
}

var _ compositeerrors.CompositeError = (*AWSPermissionError)(nil)

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

func (e *AWSPermissionError) Hints() compositeerrors.Hints {
	return compositeerrors.NewHints().WithSkipAutoRetry()
}

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

var awsPermissionPatterns = []*regexp.Regexp{
	regexp.MustCompile(
		`(?P<code>AccessDenied(?:Exception)?|AuthorizationError):\s*(?:User|Principal):\s*(?P<principal>arn:[^\s]+)\s+is not authorized to perform:\s*(?P<action>[a-zA-Z0-9-]+:[a-zA-Z0-9*]+)(?:\s+on\s+resource:\s*(?P<resource>\S+))?`,
	),
	regexp.MustCompile(
		`(?:User|Principal):\s*(?P<principal>arn:[^\s]+)\s+is not authorized to perform:\s*(?P<action>[a-zA-Z0-9-]+:[a-zA-Z0-9*]+)(?:\s+on\s+resource:\s*(?P<resource>\S+))?`,
	),
	regexp.MustCompile(
		`(?P<code>UnauthorizedOperation):[^\n]*?(?:Operation|operation):\s*(?P<action>[a-zA-Z0-9-]+:[a-zA-Z0-9*]+)`,
	),
}

func parsePermission(ctx *errparse.ParseContext) compositeerrors.CompositeError {
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
	errparse.Register(errparse.NewParser(errparse.LayerProvider, parsePermission,
		errparse.WithSignals(
			"AccessDenied",
			"UnauthorizedOperation",
			"not authorized to perform",
			"AuthorizationError",
		),
		errparse.WithProviders(errparse.ProviderAWS),
	))
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

// why: cleanResource normalises a captured resource ARN. AWS quotes the ARN in some
// messages ("... on resource: \"arn:aws:s3:::bucket\" with an explicit deny
// ...") and glues sentence punctuation to it in others; both would otherwise
// leak into the "Context" section and the copy-pasteable IAM policy statement.
func cleanResource(s string) string {
	s = trimTrailingPunct(s)
	s = strings.Trim(s, `"'`)
	return trimTrailingPunct(s)
}

func trimTrailingPunct(s string) string {
	return strings.TrimRight(s, ".,;:")
}

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
