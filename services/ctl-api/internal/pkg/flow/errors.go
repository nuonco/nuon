package flow

import "fmt"

// FlowStoppedErr is returned when a workflow is stopped due to a denial or
// skip-dependents response. Unlike ErrNotApproved, this carries context about
// why the flow stopped and signals to the execute-flow signal that it should
// not enter the retry-wait loop.
type FlowStoppedErr struct {
	StepID string
	Reason string // "denied", "skipped-dependents"
}

func (e *FlowStoppedErr) Error() string {
	return fmt.Sprintf("workflow stopped at step %s: %s", e.StepID, e.Reason)
}

func NewFlowStoppedErr(stepID, reason string) *FlowStoppedErr {
	return &FlowStoppedErr{StepID: stepID, Reason: reason}
}
