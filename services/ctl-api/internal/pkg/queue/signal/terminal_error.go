package signal

import (
	"fmt"
	"regexp"

	"go.temporal.io/sdk/temporal"
)

// TerminalErrorPrefix marks an error as a terminal step failure: retrying
// the step will not change the outcome without external intervention.
//
// The marker is encoded in the error message because Temporal and queue
// boundaries flatten everything else: Go error types are lost across activity
// boundaries, ApplicationError type and details are lost when the queue
// handler reduces the failure to a human description (see HumanError) and
// AwaitSignal rebuilds a fresh ApplicationError from the DB status.
//
// Format: "TERMINAL_ERROR[<reason_code>]: <user-facing message>"
//
// reason_code is a stable, snake_case identifier for grouping / metrics. The
// conductor reads this in pkg/flow/signals/executeworkflowstep/process_errors.go
// to short-circuit auto-retry. The marker travels through to display surfaces
// alongside the user-facing message — keeping it visible avoids any
// terminal-error-specific handling in step status / description code paths.
const TerminalErrorPrefix = "TERMINAL_ERROR"

// TerminalErrorType is the Temporal ApplicationError type used at the
// activity boundary. The type field does not survive past the queue handler
// (see await_signal.go), but it is useful while the error is still inside the
// activity-to-workflow hop.
const TerminalErrorType = "TerminalError"

// terminalErrorRe extracts the reason_code from a terminal error message.
// reason_code must be lowercase snake_case.
var terminalErrorRe = regexp.MustCompile(`TERMINAL_ERROR\[([a-z0-9_]+)\]`)

// NewTerminalError wraps a step failure as terminal with a stable
// reason_code. Terminal errors short-circuit step-level auto-retry (see
// pkg/flow/signals/executeworkflowstep/process_errors.go) so the user sees
// the failure immediately instead of after the auto-retry budget.
//
// The format/args form the user-facing message — write copy that explains
// what to do next, e.g. "Ensure there is an active build for this component
// before retrying." The TERMINAL_ERROR[<code>]: prefix is preserved when the
// message is rendered, so the copy should still read well alongside it.
//
// Use this for failures that retrying cannot fix without user action, e.g.:
//   - missing component build / build in error state
//   - missing required input
//   - invalid configuration
//
// reasonCode must match [a-z0-9_]+.
func NewTerminalError(reasonCode, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	return temporal.NewNonRetryableApplicationError(
		fmt.Sprintf("%s[%s]: %s", TerminalErrorPrefix, reasonCode, msg),
		TerminalErrorType,
		nil,
	)
}

// WrapTerminal wraps an existing error as a terminal failure, preserving the
// underlying message. Returns nil when err is nil.
func WrapTerminal(err error, reasonCode string) error {
	if err == nil {
		return nil
	}
	return NewTerminalError(reasonCode, "%s", err.Error())
}

// IsTerminalError reports whether err is — or has been wrapped as — a
// terminal error somewhere in its message chain.
func IsTerminalError(err error) bool {
	return err != nil && terminalErrorRe.MatchString(err.Error())
}

// TerminalReasonCode extracts the reason_code from a terminal error, or "" if
// err is not terminal.
func TerminalReasonCode(err error) string {
	if err == nil {
		return ""
	}
	m := terminalErrorRe.FindStringSubmatch(err.Error())
	if len(m) < 2 {
		return ""
	}
	return m[1]
}
