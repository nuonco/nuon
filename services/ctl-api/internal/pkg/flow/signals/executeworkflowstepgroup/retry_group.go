package executeworkflowstepgroup

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

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

	for _, step := range steps {
		if step.Status.Status == app.StatusDiscarded {
			continue
		}
		_ = activities.AwaitPkgWorkflowsFlowUpdateFlowStepRetried(ctx, activities.UpdateFlowStepRetriedRequest{
			StepID: step.ID,
		})
	}

	latestGroupRetryIdx := newGroupRetryIdx - 1
	var stepsToClone []app.WorkflowStep
	for _, step := range steps {
		if step.GroupRetryIdx != latestGroupRetryIdx || step.RetryIndex > 0 {
			continue
		}
		stepsToClone = append(stepsToClone, step)
	}

	groupMaxRetries := GroupMaxRetriesForSteps(stepsToClone)
	if newGroupRetryIdx > groupMaxRetries {
		l.Warn("group retry exceeds per-step max retries",
			zap.Int("new_group_retry_idx", newGroupRetryIdx),
			zap.Int("group_max_retries", groupMaxRetries))
		return fmt.Errorf("group retry %d exceeds max retries %d", newGroupRetryIdx, groupMaxRetries)
	}

	cloneSteps := make([]activities.CreateFlowStep, 0, len(stepsToClone))
	for i, step := range stepsToClone {
		cloneQueueSignal := step.QueueSignal
		if step.QueueSignal != nil && step.QueueSignal.Signal != nil {
			if cl, ok := step.QueueSignal.Signal.(signal.SignalWithClone); ok {
				defs, cloneErr := cl.Clone(ctx, step.Name)
				if cloneErr != nil {
					return errors.Wrapf(cloneErr, "unable to clone signal for retry on step %s", step.Name)
				}
				if len(defs) > 0 {
					cloneQueueSignal = &signaldb.SignalData{Signal: defs[len(defs)-1].Signal}
				}
			}
		}

		cloneSteps = append(cloneSteps, activities.CreateFlowStep{
			FlowID:      s.WorkflowID,
			OwnerID:     step.OwnerID,
			OwnerType:   step.OwnerType,
			Name:        step.Name,
			Signal:      step.Signal,
			QueueSignal: cloneQueueSignal,
			Status: app.NewCompositeTemporalStatus(ctx, app.StatusPending, map[string]any{
				"is_retry":        true,
				"retry_idx":       newGroupRetryIdx,
				"group_retry_idx": newGroupRetryIdx,
				"retry_type":      "auto",
			}),
			Idx:            maxIdx + 100*(i+1),
			ExecutionType:  step.ExecutionType,
			Metadata:       step.Metadata,
			Retryable:      step.Retryable,
			Skippable:      step.Skippable,
			SkipOnFailure:  step.SkipOnFailure,
			GroupIdx:       step.GroupIdx,
			GroupRetryIdx:  newGroupRetryIdx,
			StepTargetType: step.StepTargetType,
			RetryIndex:     0,
			Timeout:        step.Timeout,
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

func GroupMaxRetriesForSteps(steps []app.WorkflowStep) int {
	minRetries := signal.DefaultMaxRetries

	for _, step := range steps {
		if step.QueueSignal == nil || step.QueueSignal.Signal == nil {
			continue
		}
		sig := step.QueueSignal.Signal

		checkMax(sig, &minRetries)
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
