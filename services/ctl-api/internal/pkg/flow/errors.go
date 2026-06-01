package flow

import "fmt"

// ContinueAsNewErr signals that the workflow should continue-as-new from
// the given group/step index.
type ContinueAsNewErr struct {
	StartFromStepIdx int
}

func (e *ContinueAsNewErr) Error() string {
	return "continue executing this workflow as new"
}

func NewContinueAsNewErr(startsFromStepIdx int) *ContinueAsNewErr {
	return &ContinueAsNewErr{
		StartFromStepIdx: startsFromStepIdx,
	}
}

// ApprovalPauseErr indicates that execution stopped because a step is awaiting approval.
type ApprovalPauseErr struct {
	StepID string
}

func (e *ApprovalPauseErr) Error() string {
	return "workflow paused at approval step " + e.StepID
}

func NewApprovalPauseErr(stepID string) *ApprovalPauseErr {
	return &ApprovalPauseErr{StepID: stepID}
}

// FlowStoppedErr is returned when a workflow is stopped due to a denial or
// skip-dependents response. Unlike ErrNotApproved, this carries context about
// why the flow stopped and signals to the execute-flow signal that it should
// not enter the retry-wait loop.
type FlowStoppedErr struct {
	StepID           string
	Reason           string
	RetriesExhausted bool
}

func (e *FlowStoppedErr) Error() string {
	return fmt.Sprintf("workflow stopped at step %s: %s", e.StepID, e.Reason)
}

func NewFlowStoppedErr(stepID, reason string) *FlowStoppedErr {
	return &FlowStoppedErr{StepID: stepID, Reason: reason}
}
