package executeworkflowstep

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// CreateStepRetryResponse is the response from the "create-step-retry" update handler.
type CreateStepRetryResponse struct {
	// Directive is the result directive written to the step: "retry" for a
	// single-step clone, "retry-group" when the group should be retried.
	Directive string `json:"directive"`
	NewStepID string `json:"new_step_id"`
}

// createStepRetryHandler marks the step as discarded and determines the retry
// directive. The actual cloning is handled by the group which has the right
// context (WorkflowStepGroupID, etc.).
func (s *Signal) createStepRetryHandler(ctx workflow.Context) (*CreateStepRetryResponse, error) {
	step, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, s.StepID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get step")
	}
	if !step.Retryable {
		return nil, errors.New("step is not retryable")
	}

	sig := stepSignal(step)

	// Enforce the global retry ceiling — no retries past MaxRetries.
	maxRetries := signal.DefaultMaxRetries
	if mr, ok := sig.(signal.SignalWithMaxRetries); ok {
		maxRetries = mr.MaxRetries()
	}

	retryIndex := step.RetryIndex
	if rg, ok := sig.(signal.SignalWithRetryGroup); ok && rg.RetryGroup() {
		retryIndex = step.GroupRetryIdx
	}

	if retryIndex >= maxRetries {
		return nil, errors.Errorf("max retries exhausted (%d/%d)", retryIndex, maxRetries)
	}

	// Determine the directive based on signal capabilities.
	directive := DirectiveRetry
	if rg, ok := sig.(signal.SignalWithRetryGroup); ok && rg.RetryGroup() {
		directive = DirectiveRetryGroup
	}

	// Mark original step as retried.
	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepRetried(ctx, activities.UpdateFlowStepRetriedRequest{
		StepID: step.ID,
	}); err != nil {
		return nil, errors.Wrap(err, "unable to mark step as retried")
	}

	// Mark original step as discarded.
	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusDiscarded,
			StatusHumanDescription: "Step was discarded and retried.",
			Metadata: map[string]any{
				"retry_type": "manual",
				"directive":  directive,
			},
		},
	}); err != nil {
		return nil, errors.Wrap(err, "unable to mark step as discarded")
	}

	// Write the directive on the step.
	if err := setResultDirective(ctx, step.ID, directive); err != nil {
		return nil, errors.Wrap(err, "unable to set result directive")
	}

	// Unblock waitForApprovalResponse if the step is waiting for approval.
	s.retried = true

	return &CreateStepRetryResponse{Directive: directive}, nil
}
