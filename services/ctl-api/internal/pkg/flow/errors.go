package flow

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

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

type ApprovalPauseErr struct {
	StepID string
}

func (e *ApprovalPauseErr) Error() string {
	return "workflow paused at approval step " + e.StepID
}

func NewApprovalPauseErr(stepID string) *ApprovalPauseErr {
	return &ApprovalPauseErr{StepID: stepID}
}

type AwaitRetryPauseErr struct{}

func (e *AwaitRetryPauseErr) Error() string {
	return "workflow paused awaiting retry"
}

type FlowStoppedErr struct {
	StepID                 string
	Reason                 string
	RetriesExhausted       bool
	StatusHumanDescription string
	CompositeError         *compositeerrors.CompositeErrorData
}

func (e *FlowStoppedErr) Error() string {
	switch {
	case e.StepID == "" && e.Reason == "":
		return "workflow stopped"
	case e.StepID == "":
		return fmt.Sprintf("workflow stopped: %s", e.Reason)
	case e.Reason == "":
		return fmt.Sprintf("workflow stopped at step %s", e.StepID)
	default:
		return fmt.Sprintf("workflow stopped at step %s: %s", e.StepID, e.Reason)
	}
}

func NewFlowStoppedErr(stepID, reason string) *FlowStoppedErr {
	return &FlowStoppedErr{StepID: stepID, Reason: reason}
}
