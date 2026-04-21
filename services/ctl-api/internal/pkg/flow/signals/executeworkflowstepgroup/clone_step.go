package executeworkflowstepgroup

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// cloneStepForRetry fetches the step and workflow from DB, then creates a clone
// within the same group. The step should already be marked as discarded before
// this is called.
func cloneStepForRetry(ctx workflow.Context, stepID string, workflowID string) error {
	step, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, stepID)
	if err != nil {
		return fmt.Errorf("unable to get step %s: %w", stepID, err)
	}

	flw, err := activities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, workflowID)
	if err != nil {
		return fmt.Errorf("unable to get workflow %s: %w", workflowID, err)
	}

	newRetryIndex := step.RetryIndex + 1

	maxRetries := signal.DefaultMaxRetries
	if step.QueueSignal != nil && step.QueueSignal.Signal != nil {
		if mr, ok := step.QueueSignal.Signal.(signal.SignalWithMaxRetries); ok {
			maxRetries = mr.MaxRetries()
		}
	}
	if newRetryIndex > maxRetries {
		return fmt.Errorf("step %s has exceeded maximum retry count of %d", stepID, maxRetries)
	}

	_, err = activities.AwaitPkgWorkflowsFlowCreateFlowSteps(ctx, activities.CreateFlowStepsRequest{
		Steps: []activities.CreateFlowStep{
			{
				FlowID:      flw.ID,
				OwnerID:     flw.OwnerID,
				OwnerType:   flw.OwnerType,
				Name:        step.Name,
				Signal:      step.Signal,
				QueueSignal: step.QueueSignal,
				Status: app.NewCompositeTemporalStatus(ctx, app.StatusPending, map[string]any{
					"is_retry":  true,
					"retry_idx": newRetryIndex,
				}),
				Idx:                 step.Idx + 1,
				ExecutionType:       step.ExecutionType,
				Metadata:            step.Metadata,
				Retryable:           step.Retryable,
				Skippable:           step.Skippable,
				GroupIdx:            step.GroupIdx,
				GroupRetryIdx:       step.GroupRetryIdx,
				WorkflowStepGroupID: step.WorkflowStepGroupID,
				StepTargetType:      step.StepTargetType,
				RetryIndex:          newRetryIndex,
			},
		},
	})
	return err
}
