package executeworkflowstep

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// cloneWorkflowStep creates a clone of the step for retry.
// If the step's signal implements SignalWithCloneSteps, multiple steps are created
// (e.g. a plan step followed by an apply step). Otherwise a single clone is created.
// The clone gets Idx+1 so it sorts immediately after the original step.
func (s *Signal) cloneWorkflowStep(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) error {
	newRetryIndex := step.RetryIndex + 1

	maxRetries := signal.DefaultMaxRetries
	if step.QueueSignal != nil && step.QueueSignal.Signal != nil {
		if mr, ok := step.QueueSignal.Signal.(signal.SignalWithMaxRetries); ok {
			maxRetries = mr.MaxRetries()
		}
	}
	if newRetryIndex > maxRetries {
		return fmt.Errorf("step %s has exceeded maximum retry count of %d", step.ID, maxRetries)
	}

	// If the signal defines clone steps (e.g. apply signals that need a plan step first),
	// create those instead of a simple copy.
	if step.QueueSignal != nil && step.QueueSignal.Signal != nil {
		if cs, ok := step.QueueSignal.Signal.(signal.SignalWithCloneSteps); ok {
			return s.createCloneSteps(ctx, step, flw, cs, newRetryIndex)
		}
	}

	_, err := activities.AwaitPkgWorkflowsFlowCreateFlowSteps(ctx, activities.CreateFlowStepsRequest{
		Steps: []activities.CreateFlowStep{
			{
				FlowID:      flw.ID,
				OwnerID:     flw.OwnerID,
				OwnerType:   flw.OwnerType,
				Name:        getCloneStepName(step.Name),
				Signal:      step.Signal,
				QueueSignal: step.QueueSignal,
				Status: app.NewCompositeTemporalStatus(ctx, app.StatusPending, map[string]any{
					"is_retry":   true,
					"retry_idx":  newRetryIndex,
					"retry_type": "auto",
				}),
				Idx:                 step.Idx + 100,
				ExecutionType:       step.ExecutionType,
				Metadata:            step.Metadata,
				Retryable:           step.Retryable,
				Skippable:           step.Skippable,
				GroupIdx:            step.GroupIdx,
				GroupRetryIdx:       step.GroupRetryIdx,
				WorkflowStepGroupID: step.WorkflowStepGroupID,
				StepTargetType:      step.StepTargetType,
				RetryIndex:          newRetryIndex,
				StepQueueID:         step.StepQueueID,
				TargetQueueID:       step.TargetQueueID,
			},
		},
	})
	return err
}

// createCloneSteps builds multiple steps from a SignalWithCloneSteps implementation.
func (s *Signal) createCloneSteps(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow, cs signal.SignalWithCloneSteps, retryIndex int) error {
	defs := cs.CloneSteps(removeRetryFromStepName(step.Name))
	steps := make([]activities.CreateFlowStep, 0, len(defs))
	for i, def := range defs {
		steps = append(steps, activities.CreateFlowStep{
			FlowID:      flw.ID,
			OwnerID:     flw.OwnerID,
			OwnerType:   flw.OwnerType,
			Name:        getCloneStepName(def.Name),
			QueueSignal: &signaldb.SignalData{Signal: def.Signal},
			Status: app.NewCompositeTemporalStatus(ctx, app.StatusPending, map[string]any{
				"is_retry":  true,
				"retry_idx": retryIndex,
			}),
			Idx:                 step.Idx + 100 + i,
			ExecutionType:       app.WorkflowStepExecutionType(def.ExecutionType),
			Metadata:            step.Metadata,
			Retryable:           step.Retryable,
			Skippable:           step.Skippable,
			GroupIdx:            step.GroupIdx,
			GroupRetryIdx:       step.GroupRetryIdx,
			WorkflowStepGroupID: step.WorkflowStepGroupID,
			StepTargetType:      step.StepTargetType,
			RetryIndex:          retryIndex,
			StepQueueID:         step.StepQueueID,
			TargetQueueID:       step.TargetQueueID,
		})
	}
	_, err := activities.AwaitPkgWorkflowsFlowCreateFlowSteps(ctx, activities.CreateFlowStepsRequest{Steps: steps})
	return err
}

func getCloneStepName(name string) string {
	re := regexp.MustCompile(`^(.*)\(retry (\d+)\)$`)
	matches := re.FindStringSubmatch(name)

	if len(matches) == 3 {
		base := strings.TrimSpace(matches[1])
		retryCount, err := strconv.Atoi(matches[2])
		if err == nil {
			return fmt.Sprintf("%s (retry %d)", base, retryCount+1)
		}
	}

	return fmt.Sprintf("%s (retry 1)", name)
}

func removeRetryFromStepName(name string) string {
	re := regexp.MustCompile(`^(.*)\(retry \d+\)$`)
	matches := re.FindStringSubmatch(name)
	if len(matches) == 2 {
		return strings.TrimSpace(matches[1])
	}
	return name
}
