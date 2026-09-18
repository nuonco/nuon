package executeworkflowstep

import (
	stderrors "errors"
	"regexp"
	"strings"

	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

const stepDescriptionMaxLen = 512

var (
	temporalErrorMetadataPattern = regexp.MustCompile(`\s*\((?:type|activityType|workflowType):[^()]*\)`)
	temporalErrorPrefixPattern   = regexp.MustCompile(`(?i)^(?:activity|child workflow execution|workflow execution)\s+error\s*:?\s*`)
)

// stepHumanDescription returns a user-facing error message for a failed step.
// Duplicated from flow.StepHumanDescription to avoid import cycle.
func stepHumanDescription(err error) string {
	var appErr *temporal.ApplicationError
	if stderrors.As(err, &appErr) && appErr.NonRetryable() {
		if message := sanitizeStepDescription(appErr.Message()); message != "" {
			return message
		}
	}
	return "Step failed"
}

func abandonedHumanDescription(err error) string {
	desc := stepHumanDescription(err)
	if desc == "" || desc == "Step failed" {
		return "step abandoned after failure: no retry or skip received"
	}
	return "step abandoned after failure: " + desc
}

// sanitizeStepDescription strips Temporal's nested error chain metadata and any
// credentials before a raw message is shown to a user.
func sanitizeStepDescription(message string) string {
	cleaned := temporalErrorMetadataPattern.ReplaceAllString(message, "")
	cleaned = compositeerrors.RedactDiagnosticSecrets(cleaned)
	for _, line := range strings.Split(cleaned, "\n") {
		line = strings.TrimSpace(temporalErrorPrefixPattern.ReplaceAllString(strings.TrimSpace(line), ""))
		if line == "" {
			continue
		}
		return truncateRunes(line, stepDescriptionMaxLen)
	}
	return ""
}

func truncateRunes(s string, n int) string {
	if len(s) <= n {
		return s
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}
