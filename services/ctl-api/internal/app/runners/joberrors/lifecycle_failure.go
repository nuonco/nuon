package joberrors

import "github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"

const LifecycleFailureErrorType compositeerrors.Type = "runner.job_lifecycle_failure"

type LifecycleFailureReason string

const (
	LifecycleFailureReasonNoActiveRunner    LifecycleFailureReason = "no_active_runner"
	LifecycleFailureReasonQueueTimeout      LifecycleFailureReason = "queue_timeout"
	LifecycleFailureReasonRunnerUnhealthy   LifecycleFailureReason = "runner_unhealthy"
	LifecycleFailureReasonPickupTimeout     LifecycleFailureReason = "pickup_timeout"
	LifecycleFailureReasonOverallTimeout    LifecycleFailureReason = "overall_timeout"
	LifecycleFailureReasonExecutionTimeout  LifecycleFailureReason = "execution_timeout"
	LifecycleFailureReasonAttemptsExhausted LifecycleFailureReason = "attempts_exhausted"
	LifecycleFailureReasonResultMissing     LifecycleFailureReason = "execution_result_missing"
	LifecycleFailureReasonRunnerDisabled    LifecycleFailureReason = "runner_disabled"
)

type LifecycleFailureError struct {
	Reason LifecycleFailureReason `json:"reason"`
}

var _ compositeerrors.CompositeError = (*LifecycleFailureError)(nil)

func (e *LifecycleFailureError) Error() string {
	switch e.Reason {
	case LifecycleFailureReasonNoActiveRunner:
		return "No active runner was available for this job"
	case LifecycleFailureReasonRunnerDisabled:
		return "The install runner is disabled"
	case LifecycleFailureReasonQueueTimeout:
		return "Runner job expired in the queue"
	case LifecycleFailureReasonRunnerUnhealthy:
		return "Runner was unavailable while processing this job"
	case LifecycleFailureReasonPickupTimeout:
		return "Runner did not pick up the job in time"
	case LifecycleFailureReasonOverallTimeout:
		return "Runner job exceeded its overall timeout"
	case LifecycleFailureReasonExecutionTimeout:
		return "Runner job execution timed out"
	case LifecycleFailureReasonAttemptsExhausted:
		return "Runner job exhausted its execution attempts"
	case LifecycleFailureReasonResultMissing:
		return "Runner did not report a result for this job"
	default:
		return "Runner job could not be completed"
	}
}

func (*LifecycleFailureError) Type() compositeerrors.Type {
	return LifecycleFailureErrorType
}

// Hints marks a disabled runner as terminal: nothing about the failure can
// change until the customer re-applies their stack, so neither auto-retries nor
// a manual retry can succeed.
func (e *LifecycleFailureError) Hints() compositeerrors.Hints {
	if e.Reason == LifecycleFailureReasonRunnerDisabled {
		return compositeerrors.NewHints().WithTerminal()
	}
	return nil
}

func (*LifecycleFailureError) Severity() compositeerrors.Severity {
	return compositeerrors.SeverityError
}

func (e *LifecycleFailureError) Sections() []compositeerrors.Section {
	why, fix := e.details()
	return []compositeerrors.Section{
		compositeerrors.MarkdownSection("What happened", why),
		compositeerrors.MarkdownSection("How to fix", fix),
	}
}

func (e *LifecycleFailureError) details() (string, string) {
	switch e.Reason {
	case LifecycleFailureReasonNoActiveRunner:
		return "The job could not start because its assigned runner had no active process.", "Check that the runner is online and healthy, then retry the operation."
	case LifecycleFailureReasonRunnerDisabled:
		return "The install runner is disabled, so no runner exists to pick up this job.", "Re-enable the runner in the install stack, then retry the operation."
	case LifecycleFailureReasonQueueTimeout:
		return "The job waited in the queue longer than its configured queue timeout.", "Check the runner's health and workload, then retry the operation."
	case LifecycleFailureReasonRunnerUnhealthy:
		return "The runner was not healthy long enough to start or continue the job.", "Restore the runner to a healthy state, then retry the operation."
	case LifecycleFailureReasonPickupTimeout:
		return "The job was made available, but the runner did not reserve it before the pickup timeout.", "Check the runner's connectivity and health, then retry the operation."
	case LifecycleFailureReasonOverallTimeout:
		return "The job did not complete before its overall timeout expired.", "Check the runner and the operation's progress, then retry once the underlying issue is resolved."
	case LifecycleFailureReasonExecutionTimeout:
		return "The runner reserved the job, but the execution did not finish before its timeout.", "Inspect the runner logs for the operation, address anything preventing completion, then retry."
	case LifecycleFailureReasonAttemptsExhausted:
		return "Every available execution attempt ended before the job completed.", "Check the runner's health and logs, then retry the operation."
	case LifecycleFailureReasonResultMissing:
		return "The execution ended, but the runner did not report its result details.", "Check the runner's health and logs, then retry the operation."
	default:
		return "The runner job ended before it could complete.", "Check the runner's health and logs, then retry the operation."
	}
}
