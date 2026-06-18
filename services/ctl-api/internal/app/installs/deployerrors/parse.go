package deployerrors

import (
	"regexp"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

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

// Parse inspects raw terraform plan/apply error output and returns a typed
// AWSPermissionError when it recognises an AWS IAM permission failure. It
// returns nil when there is no confident match, so callers fall back to the
// existing plain-string status description.
func Parse(raw string) compositeerrors.CompositeError {
	if !strings.Contains(raw, "AccessDenied") &&
		!strings.Contains(raw, "UnauthorizedOperation") &&
		!strings.Contains(raw, "not authorized to perform") &&
		!strings.Contains(raw, "AuthorizationError") {
		return nil
	}

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
			Resource:     trimTrailingPunct(fields["resource"]),
			Principal:    fields["principal"],
			AWSErrorCode: fields["code"],
			RawMessage:   extractRelevantLine(raw, match[0]),
		}
	}

	return nil
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
