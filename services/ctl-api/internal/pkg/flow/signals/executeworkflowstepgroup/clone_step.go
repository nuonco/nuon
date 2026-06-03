package executeworkflowstepgroup

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// cloneStepForRetry fetches the step and workflow from DB, then creates a clone
// within the same group. The step should already be marked as discarded before
// this is called.
//
// If the step's signal implements SignalWithClone, Clone() is called to produce
// one or more replacement steps (e.g., a plan signal returns a clean copy, an
// apply signal returns plan + apply). Otherwise the signal is copied verbatim.
func CloneStepForRetry(ctx workflow.Context, stepID string, workflowID string) error {
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

	// If the signal implements Clone(), use it to produce replacement steps.
	if step.QueueSignal != nil && step.QueueSignal.Signal != nil {
		if cl, ok := step.QueueSignal.Signal.(signal.SignalWithClone); ok {
			defs, cloneErr := cl.Clone(ctx, step.Name)
			if cloneErr != nil {
				return fmt.Errorf("unable to clone signal for retry: %w", cloneErr)
			}
			return createCloneStepsFromDefs(ctx, step, flw, defs, newRetryIndex)
		}
	}

	// Default: copy signal verbatim as a single step.
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
				Timeout:             step.Timeout,
			},
		},
	})
	return err
}

// createCloneStepsFromDefs builds workflow steps from Clone()-returned defs.
func createCloneStepsFromDefs(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow, defs []signal.CloneStepDef, retryIndex int) error {
	steps := make([]activities.CreateFlowStep, 0, len(defs))
	for i, def := range defs {
		execType := step.ExecutionType
		if def.ExecutionType != "" {
			execType = app.WorkflowStepExecutionType(def.ExecutionType)
		}
		name := def.Name
		if name == "" {
			name = step.Name
		}

		steps = append(steps, activities.CreateFlowStep{
			FlowID:      flw.ID,
			OwnerID:     flw.OwnerID,
			OwnerType:   flw.OwnerType,
			Name:        name,
			Signal:      step.Signal,
			QueueSignal: &signaldb.SignalData{Signal: def.Signal},
			Status: app.NewCompositeTemporalStatus(ctx, app.StatusPending, map[string]any{
				"is_retry":  true,
				"retry_idx": retryIndex,
			}),
			Idx:                 step.Idx + 1 + i,
			ExecutionType:       execType,
			Metadata:            step.Metadata,
			Retryable:           step.Retryable,
			Skippable:           step.Skippable,
			GroupIdx:            step.GroupIdx,
			GroupRetryIdx:       step.GroupRetryIdx,
			WorkflowStepGroupID: step.WorkflowStepGroupID,
			StepTargetType:      step.StepTargetType,
			RetryIndex:          retryIndex,
			Timeout:             step.Timeout,
		})
	}

	_, err := activities.AwaitPkgWorkflowsFlowCreateFlowSteps(ctx, activities.CreateFlowStepsRequest{
		Steps: steps,
	})
	return err
}
