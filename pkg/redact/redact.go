package redact

import "regexp"

const Placeholder = "[REDACTED]"

const sensitiveCredentialName = `(?:access[-_]?key|access[-_]?token|api[-_]?key|auth[-_]?token|awsaccesskeyid|client[-_]?assertion|client[-_]?secret|id[-_]?token|password|refresh[-_]?token|secret|session[-_]?token|sig|signature|token|x-amz-credential|x-amz-security-token|x-amz-signature|x-goog-credential|x-goog-signature)`
const diagnosticCredentialName = `(?:[[:alnum:]]+[-_.])*` + sensitiveCredentialName

type diagnosticSecretRedactor struct {
	pattern     *regexp.Regexp
	replacement string
}

var diagnosticSecretRedactors = []diagnosticSecretRedactor{
	{
		// Credentials embedded in URL userinfo, such as
		// https://user:password@host or postgres://user:password@host.
		pattern:     regexp.MustCompile(`(?i)(\b(?:https?|ssh|git|postgres(?:ql)?|mysql|redis|mongodb(?:\+srv)?|mssql|sqlserver|amqps?)://)[^/@\s]+@`),
		replacement: "${1}" + Placeholder + "@",
	},
	{
		// HTTP authorization and cookie header values. Line-anchored so a wrapped
		// "unable to get authorization: ..." keeps its cause chain.
		pattern:     regexp.MustCompile(`(?im)^([ \t]*(?:authorization|proxy-authorization|cookie|set-cookie)[ \t]*:[ \t]*)[^\r\n]*`),
		replacement: "${1}" + Placeholder,
	},
	{
		// Quoted JSON-style credential fields, such as "api_token":"...".
		pattern:     regexp.MustCompile(`(?i)(["']` + diagnosticCredentialName + `["']\s*:\s*["'])[^"'\r\n]*`),
		replacement: "${1}" + Placeholder,
	},
	{
		// Quoted HCL, shell, or environment assignments, such as password = "...".
		pattern:     regexp.MustCompile(`(?i)(\b` + diagnosticCredentialName + `\s*=\s*["'])[^"'\r\n]*`),
		replacement: "${1}" + Placeholder,
	},
	{
		// Unquoted URL query and environment assignments, such as token=<value>.
		// The colon form requires no whitespace after it so prose following a
		// wrapped "word: " is kept intact.
		pattern:     regexp.MustCompile(`(?i)(\b` + diagnosticCredentialName + `(?:\s*=\s*|:))[^&\s"'<>]+`),
		replacement: "${1}" + Placeholder,
	},
	{
		pattern:     regexp.MustCompile(`(?i)(\bBearer\s+)[a-z0-9._~+/-]+=*`),
		replacement: "${1}" + Placeholder,
	},
	{
		pattern:     regexp.MustCompile(`\beyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+\b`),
		replacement: Placeholder,
	},
}

var credentialKey = regexp.MustCompile(`(?i)^` + diagnosticCredentialName + `$`)

// DiagnosticSecrets removes credentials in common URL, header, and
// structured-assignment forms. It is intentionally not a general-purpose
// detector for arbitrary secrets embedded in otherwise unlabeled prose.
func DiagnosticSecrets(value string) string {
	for _, redactor := range diagnosticSecretRedactors {
		value = redactor.pattern.ReplaceAllString(value, redactor.replacement)
	}
	return value
}

func IsCredentialKey(key string) bool {
	return credentialKey.MatchString(key)
}
