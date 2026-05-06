package aws_missing_iam_permission

import (
	"context"
	"regexp"
	"strings"

	composite_error "github.com/nuonco/nuon/pkg/composite_error"
)

const (
	parserName    = "aws.missing_iam_permission"
	parserVersion = "1"
)

// Parser extracts AccessDenied / UnauthorizedOperation errors from any AWS-
// touching context: terraform, helm (when terraform is the IaC backing it),
// and runner-job output.
//
// The parser is conservative: it only matches when it can extract the IAM
// action. If the action isn't recoverable, it returns Matched=false so the
// broader parser (e.g. terraform_apply_failed) handles the input.
type Parser struct{}

var _ composite_error.Parser = (*Parser)(nil)

func (Parser) Name() string    { return parserName }
func (Parser) Version() string { return parserVersion }

// Contexts: register against every subtree where AWS API calls happen.
// Adding a new subtree (e.g. "build/terraform") is just a one-line addition
// here — no need to re-register elsewhere.
func (Parser) Contexts() []composite_error.ParseContext {
	return []composite_error.ParseContext{
		"terraform",
		"helm",
		"runner/job",
	}
}

// Patterns we attempt, in order. Each must define named groups for "action",
// and may define "principal" / "resource" / "code".
var awsPatterns = []*regexp.Regexp{
	// Classic AccessDenied with principal + action + resource.
	// e.g. "User: arn:aws:iam::123456789012:user/foo is not authorized to perform: ec2:CreateVpc on resource: arn:aws:ec2:..."
	regexp.MustCompile(
		`(?P<code>AccessDenied(?:Exception)?|AuthorizationError):\s*(?:User|Principal):\s*(?P<principal>arn:[^\s]+)\s+is not authorized to perform:\s*(?P<action>[a-zA-Z0-9-]+:[a-zA-Z0-9*]+)(?:\s+on\s+resource:\s*(?P<resource>\S+))?`,
	),
	// Same shape without explicit AccessDenied prefix (some SDK clients).
	regexp.MustCompile(
		`(?:User|Principal):\s*(?P<principal>arn:[^\s]+)\s+is not authorized to perform:\s*(?P<action>[a-zA-Z0-9-]+:[a-zA-Z0-9*]+)(?:\s+on\s+resource:\s*(?P<resource>\S+))?`,
	),
	// EC2-style UnauthorizedOperation. Action lives in a separate sentence.
	// e.g. "UnauthorizedOperation: You are not authorized to perform this operation. ... Operation: ec2:DescribeVpcs"
	regexp.MustCompile(
		`(?P<code>UnauthorizedOperation):[^\n]*?(?:Operation|operation):\s*(?P<action>[a-zA-Z0-9-]+:[a-zA-Z0-9*]+)`,
	),
}

func (p Parser) Parse(_ context.Context, in composite_error.ParseInput) composite_error.ParseResult {
	raw := string(in.Raw)
	if !strings.Contains(raw, "AccessDenied") &&
		!strings.Contains(raw, "UnauthorizedOperation") &&
		!strings.Contains(raw, "not authorized to perform") &&
		!strings.Contains(raw, "AuthorizationError") {
		return composite_error.ParseResult{Matched: false}
	}

	for _, re := range awsPatterns {
		match := re.FindStringSubmatch(raw)
		if match == nil {
			continue
		}
		fields := groupMap(re, match)
		action := fields["action"]
		if action == "" {
			continue
		}

		e := &Error{
			Action:       action,
			Resource:     trimTrailingPunct(fields["resource"]),
			Principal:    fields["principal"],
			AWSErrorCode: fields["code"],
			RawMessage:   extractRelevantLine(raw, match[0]),
		}

		return composite_error.ParseResult{
			Matched: true,
			Error:   e,
			Source: composite_error.Source{
				ParserName:    parserName,
				ParserVersion: parserVersion,
				Snippet:       composite_error.CapSnippet(raw),
			},
		}
	}

	return composite_error.ParseResult{Matched: false}
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

// trimTrailingPunct strips trailing punctuation that often glues to ARNs
// when AWS embeds them in sentences.
func trimTrailingPunct(s string) string {
	return strings.TrimRight(s, ".,;:")
}

// extractRelevantLine returns the line containing matchStr, trimmed of the
// terraform "│ " prefix when present.
func extractRelevantLine(raw, matchStr string) string {
	for _, line := range strings.Split(raw, "\n") {
		if strings.Contains(line, firstFiftyChars(matchStr)) {
			t := strings.TrimSpace(line)
			t = strings.TrimPrefix(t, "│")
			return strings.TrimSpace(t)
		}
	}
	return strings.TrimSpace(matchStr)
}

func firstFiftyChars(s string) string {
	if len(s) <= 50 {
		return s
	}
	return s[:50]
}
