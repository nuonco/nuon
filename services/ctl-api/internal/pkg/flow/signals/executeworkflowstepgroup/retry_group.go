package executeworkflowstepgroup

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// retryGroup clones all steps in the group for retry. This is the core RetryGroup
// logic — it lives on the group signal, not on the step signal.
//
// 1. Fetches all steps in the group
// 2. Guards against parallel groups (retry-group not supported)
// 3. Marks all steps as Discarded
// 4. Clones each step (respecting SignalWithCloneSteps)
// 5. Increments GroupRetryIdx on all clones
func (s *Signal) retryGroup(ctx workflow.Context, l *zap.Logger) error {
	if s.Parallel {
		return errors.New("retry-group is not supported for parallel groups")
	}

	steps, err := s.getGroupSteps(ctx)
	if err != nil {
		return err
	}

	if len(steps) == 0 {
		return errors.New("no steps found in group to retry")
	}

	// Determine the new GroupRetryIdx and base Idx for clones
	maxIdx := 0
	newGroupRetryIdx := 0
	for _, step := range steps {
		if step.Idx > maxIdx {
			maxIdx = step.Idx
		}
		if step.GroupRetryIdx >= newGroupRetryIdx {
			newGroupRetryIdx = step.GroupRetryIdx + 1
		}
	}

	l.Debug("retrying group",
		zap.Int("group_idx", s.GroupIdx),
		zap.Int("step_count", len(steps)),
		zap.Int("new_group_retry_idx", newGroupRetryIdx))

	// Mark ALL steps in the group as retried and discarded (including terminal
	// ones like StatusSuccess) so the dashboard clearly shows which generation
	// is active.
	for _, step := range steps {
		if step.Status.Status == app.StatusDiscarded {
			continue
		}
		_ = activities.AwaitPkgWorkflowsFlowUpdateFlowStepRetried(ctx, activities.UpdateFlowStepRetriedRequest{
			StepID: step.ID,
		})
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status:                 app.StatusDiscarded,
				StatusHumanDescription: "Group was retried.",
				Metadata: map[string]any{
					"reason": "group retry",
				},
			},
		}); err != nil {
			l.Warn("failed to mark step as discarded", zap.String("step_id", step.ID), zap.Error(err))
		}
	}

	// Clone each step
	cloneSteps := make([]activities.CreateFlowStep, 0, len(steps))
	for i, step := range steps {
		// Check if the signal defines custom clone steps
		if step.QueueSignal != nil && step.QueueSignal.Signal != nil {
			if cs, ok := step.QueueSignal.Signal.(signal.SignalWithCloneSteps); ok {
				defs := cs.CloneSteps(step.Name)
				for j, def := range defs {
					cloneSteps = append(cloneSteps, activities.CreateFlowStep{
						FlowID:      s.WorkflowID,
						OwnerID:     step.OwnerID,
						OwnerType:   step.OwnerType,
						Name:        fmt.Sprintf("%s (retry %d)", def.Name, newGroupRetryIdx),
						QueueSignal: &signaldb.SignalData{Signal: def.Signal},
						Status: app.NewCompositeTemporalStatus(ctx, app.StatusPending, map[string]any{
							"is_retry":        true,
							"retry_idx":       0,
							"group_retry_idx": newGroupRetryIdx,
						}),
						Idx:            maxIdx + 100 + (i * 10) + j,
						ExecutionType:  app.WorkflowStepExecutionType(def.ExecutionType),
						Metadata:       step.Metadata,
						Retryable:      step.Retryable,
						Skippable:      step.Skippable,
						GroupIdx:       step.GroupIdx,
						GroupRetryIdx:  newGroupRetryIdx,
						StepTargetType: step.StepTargetType,
						RetryIndex:     0,
					})
				}
				continue
			}
		}

		// Simple clone
		cloneSteps = append(cloneSteps, activities.CreateFlowStep{
			FlowID:      s.WorkflowID,
			OwnerID:     step.OwnerID,
			OwnerType:   step.OwnerType,
			Name:        fmt.Sprintf("%s (retry %d)", step.Name, newGroupRetryIdx),
			Signal:      step.Signal,
			QueueSignal: step.QueueSignal,
			Status: app.NewCompositeTemporalStatus(ctx, app.StatusPending, map[string]any{
				"is_retry":        true,
				"retry_idx":       0,
				"group_retry_idx": newGroupRetryIdx,
			}),
			Idx:            maxIdx + 100 + i,
			ExecutionType:  step.ExecutionType,
			Metadata:       step.Metadata,
			Retryable:      step.Retryable,
			Skippable:      step.Skippable,
			GroupIdx:       step.GroupIdx,
			GroupRetryIdx:  newGroupRetryIdx,
			StepTargetType: step.StepTargetType,
			RetryIndex:     0,
		})
	}

	if len(cloneSteps) > 0 {
		if _, err := activities.AwaitPkgWorkflowsFlowCreateFlowSteps(ctx, activities.CreateFlowStepsRequest{
			Steps: cloneSteps,
		}); err != nil {
			return errors.Wrap(err, "unable to create retry group clone steps")
		}
	}

	return nil
}
