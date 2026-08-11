package joberrors

import "github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"

const CancellationErrorType compositeerrors.Type = "runner.job_cancelled"

type CancellationReason string

const (
	CancellationReasonAPI               CancellationReason = "api"
	CancellationReasonAttemptsExhausted CancellationReason = "attempts_exhausted"
)

type CancellationError struct {
	Reason CancellationReason `json:"reason"`
}

var _ compositeerrors.CompositeError = (*CancellationError)(nil)

func (e *CancellationError) Error() string {
	if e.Reason == CancellationReasonAttemptsExhausted {
		return "Runner job was cancelled after exhausting its execution attempts"
	}
	return "Runner job was cancelled through the API"
}

func (*CancellationError) Type() compositeerrors.Type {
	return CancellationErrorType
}

func (*CancellationError) Severity() compositeerrors.Severity {
	return compositeerrors.SeverityWarning
}

func (e *CancellationError) Sections() []compositeerrors.Section {
	if e.Reason == CancellationReasonAttemptsExhausted {
		return []compositeerrors.Section{
			compositeerrors.MarkdownSection("What happened", "The job reached its maximum number of execution attempts without completing."),
			compositeerrors.MarkdownSection("How to continue", "Check the runner and job logs, resolve the underlying failure, then retry the operation."),
		}
	}
	return []compositeerrors.Section{
		compositeerrors.MarkdownSection("What happened", "The job was cancelled before it completed."),
		compositeerrors.MarkdownSection("How to continue", "Retry the operation if it still needs to run."),
	}
}
