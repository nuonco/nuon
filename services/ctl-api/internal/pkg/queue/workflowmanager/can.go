package workflowmanager

import "go.temporal.io/sdk/workflow"

type CANHintChecker interface {
	CheckCANHint(ctx workflow.Context) (bool, error)

	ClearCANHint(ctx workflow.Context) error
}

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
