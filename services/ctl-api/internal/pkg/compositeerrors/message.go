package compositeerrors

import (
	"errors"
	"regexp"
	"strings"

	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/pkg/redact"
)

const FailureMessageMaxLen = 512

var (
	temporalErrorMetadataPattern = regexp.MustCompile(`\s*\((?:type|activityType|workflowType):[^()]*\)`)
	temporalErrorPrefixPattern   = regexp.MustCompile(`(?i)^(?:activity|child workflow execution|workflow execution)\s+error\s*:?\s*`)
)

func TruncateRunes(s string, n int) string {
	if len(s) <= n {
		return s
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}

func ConciseMessage(raw string) string {
	cleaned := temporalErrorMetadataPattern.ReplaceAllString(raw, "")
	cleaned = redact.DiagnosticSecrets(cleaned)
	for _, line := range strings.Split(cleaned, "\n") {
		line = strings.TrimSpace(temporalErrorPrefixPattern.ReplaceAllString(strings.TrimSpace(line), ""))
		if line == "" {
			continue
		}
		return TruncateRunes(line, FailureMessageMaxLen)
	}
	return ""
}

func ErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	var applicationErr *temporal.ApplicationError
	if errors.As(err, &applicationErr) {
		msg = applicationErr.Message()
	}
	if message := ConciseMessage(msg); message != "" {
		return message
	}
	return TruncateRunes(redact.DiagnosticSecrets(msg), FailureMessageMaxLen)
}
