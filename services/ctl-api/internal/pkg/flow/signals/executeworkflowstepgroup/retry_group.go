package executeworkflowstepgroup

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
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

	// Mark ALL steps in the group as retried. Keep the existing status so
	// errors still show correctly in the dashboard — just set the retried
	// boolean and add group_retry metadata.
	for _, step := range steps {
		if step.Status.Status == app.StatusDiscarded {
			continue
		}
		_ = activities.AwaitPkgWorkflowsFlowUpdateFlowStepRetried(ctx, activities.UpdateFlowStepRetriedRequest{
			StepID: step.ID,
		})
	}

	// Only clone primary steps — filter out any step that is a retry clone
	// (RetryIndex > 0 from auto-retry, or GroupRetryIdx > 0 from group retry).
	// This gives us exactly the original set of steps from generation 0.
	var stepsToClone []app.WorkflowStep
	for _, step := range steps {
		if step.RetryIndex > 0 || step.GroupRetryIdx > 0 {
			continue
		}
		stepsToClone = append(stepsToClone, step)
	}

	// Enforce per-step max retries at the group level. The group retry count
	// is bounded by the minimum MaxRetries across all signals in the group
	// (including signals produced by CloneSteps). For example, if plan has
	// MaxRetries=3 and apply has MaxRetries=5, the group stops after 3.
	groupMaxRetries := groupMaxRetriesForSteps(stepsToClone)
	if newGroupRetryIdx > groupMaxRetries {
		l.Warn("group retry exceeds per-step max retries",
			zap.Int("new_group_retry_idx", newGroupRetryIdx),
			zap.Int("group_max_retries", groupMaxRetries))
		return fmt.Errorf("group retry %d exceeds max retries %d", newGroupRetryIdx, groupMaxRetries)
	}

	// Clone each step. Always simple-clone from gen 0 — CloneSteps expansion
	// was already applied during initial workflow creation, so the gen 0 steps
	// already represent the full set (e.g. plan + apply as separate steps).
	cloneSteps := make([]activities.CreateFlowStep, 0, len(stepsToClone))
	for i, step := range stepsToClone {
		cloneSteps = append(cloneSteps, activities.CreateFlowStep{
			FlowID:      s.WorkflowID,
			OwnerID:     step.OwnerID,
			OwnerType:   step.OwnerType,
			Name:        step.Name,
			Signal:      step.Signal,
			QueueSignal: step.QueueSignal,
			Status: app.NewCompositeTemporalStatus(ctx, app.StatusPending, map[string]any{
				"is_retry":        true,
				"retry_idx":       0,
				"group_retry_idx": newGroupRetryIdx,
				"retry_type":      "auto",
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

// groupMaxRetriesForSteps returns the minimum MaxRetries across all signals
// in the given steps, including signals produced by CloneSteps. Falls back to
// signal.DefaultMaxRetries for signals that don't implement SignalWithMaxRetries.
func groupMaxRetriesForSteps(steps []app.WorkflowStep) int {
	minRetries := signal.DefaultMaxRetries

	for _, step := range steps {
		if step.QueueSignal == nil || step.QueueSignal.Signal == nil {
			continue
		}
		sig := step.QueueSignal.Signal

		// Check the step's own signal.
		checkMax(sig, &minRetries)

		// If the signal produces clone steps, check those signals too.
		if cs, ok := sig.(signal.SignalWithCloneSteps); ok {
			for _, def := range cs.CloneSteps(step.Name) {
				if def.Signal != nil {
					checkMax(def.Signal, &minRetries)
				}
			}
		}
	}

	return minRetries
}

func checkMax(sig signal.Signal, minRetries *int) {
	mr, ok := sig.(signal.SignalWithMaxRetries)
	if !ok {
		return
	}
	if v := mr.MaxRetries(); v < *minRetries {
		*minRetries = v
	}
}
