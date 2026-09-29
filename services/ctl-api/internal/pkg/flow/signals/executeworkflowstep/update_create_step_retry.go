package executeworkflowstep

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

type CreateStepRetryResponse struct {
	Directive string `json:"directive"`
	NewStepID string `json:"new_step_id"`
}

func (s *Signal) createStepRetryHandler(ctx workflow.Context) (*CreateStepRetryResponse, error) {
	step, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, s.StepID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get step")
	}
	if !step.Retryable {
		return nil, temporal.NewNonRetryableApplicationError("step is not retryable", "STEP_NOT_RETRYABLE", nil)
	}

	// why: The directive is the last write below, so its presence alongside Retried
	// means an earlier landing of this update already finished; a re-delivery
	// (activity retry, double-click) must not re-run OnRetry or re-discard.
	if prior := directive.Step(step.ResultDirective); step.Retried &&
		(prior == DirectiveRetry || prior == DirectiveRetryGroup) {
		s.retried = true
		return &CreateStepRetryResponse{Directive: string(prior)}, nil
	}

	sig := stepSignal(step)

	maxRetries := signal.DefaultMaxRetries
	if mr, ok := sig.(signal.SignalWithMaxRetries); ok {
		maxRetries = mr.MaxRetries()
	}

	retryIndex := step.RetryIndex
	if rg, ok := sig.(signal.SignalWithRetryGroup); ok && rg.RetryGroup() {
		retryIndex = step.GroupRetryIdx
	}

	if retryIndex >= maxRetries {
		return nil, temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("max retries exhausted (%d/%d)", retryIndex, maxRetries),
			"MAX_RETRIES_EXHAUSTED",
			nil,
		)
	}

	directive := DirectiveRetry
	if rg, ok := sig.(signal.SignalWithRetryGroup); ok && rg.RetryGroup() {
		directive = DirectiveRetryGroup
	}

	if or, ok := sig.(signal.SignalWithOnRetry); ok {
		if err := or.OnRetry(ctx); err != nil {
			return nil, errors.Wrap(err, "OnRetry hook failed")
		}
	}

	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepRetried(ctx, activities.UpdateFlowStepRetriedRequest{
		StepID: step.ID,
	}); err != nil {
		return nil, errors.Wrap(err, "unable to mark step as retried")
	}

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

	if err := setResultDirective(ctx, step.ID, directive); err != nil {
		return nil, errors.Wrap(err, "unable to set result directive")
	}

	s.retried = true

	return &CreateStepRetryResponse{Directive: string(directive)}, nil
}
