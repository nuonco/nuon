package workflowmanager

import "go.temporal.io/sdk/workflow"

// CANHintChecker checks for externally-requested continue-as-new hints.
// Implementations typically check a metadata flag in the database.
type CANHintChecker interface {
	// CheckCANHint returns true if a continue-as-new has been requested.
	CheckCANHint(ctx workflow.Context) (bool, error)

	// ClearCANHint clears the hint so subsequent checks don't re-trigger.
	ClearCANHint(ctx workflow.Context) error
}

// CANHintCheckerFunc adapts a pair of functions into a CANHintChecker.
type CANHintCheckerFunc struct {
	CheckFn func(ctx workflow.Context) (bool, error)
	ClearFn func(ctx workflow.Context) error
}

func (f CANHintCheckerFunc) CheckCANHint(ctx workflow.Context) (bool, error) {
	return f.CheckFn(ctx)
}

func (f CANHintCheckerFunc) ClearCANHint(ctx workflow.Context) error {
	return f.ClearFn(ctx)
}
