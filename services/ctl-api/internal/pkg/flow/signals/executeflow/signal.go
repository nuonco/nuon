package executeflow

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	qsignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

const SignalType qsignal.SignalType = "execute-workflow"

type Signal struct {
	// WorkflowID is the ID of the workflow to execute.
	WorkflowID string `json:"workflow_id"`

	// Conductor configuration — set by the creator when enqueuing.
	StepGroupQueueName  string `json:"step_group_queue_name"`
	StepQueueName       string `json:"step_queue_name"`
	StepTargetQueueName string `json:"step_target_queue_name"`
	OwnerID             string `json:"owner_id"`
	OwnerType           string `json:"owner_type"`

	// Resume state — set by update handlers (approve/retry/skip) to wake the
	// main execute loop when it is waiting after an approval pause or error.
	resumeRequested bool
	resumeRunType   app.WorkflowRunType
	resumeStepID    string
	resumeStartIdx  int

	// Cancel state — set by cancel update handlers.
	cancelRequested bool

	// activeGroupQueueSignalID is the queue signal ID of the currently
	// executing group. Set by executeGroup, used by cancelWorkflowHandler
	// to actively cancel the running group.
	activeGroupQueueSignalID string

	// Pause state — set by "pause-workflow" update handler. When true, the
	// flow will pause after the current group completes.
	pauseRequested bool
}

var _ qsignal.Signal = (*Signal)(nil)
var _ qsignal.SignalWithUpdateHandlers = (*Signal)(nil)

func (s *Signal) Type() qsignal.SignalType  { return SignalType }
func (s *Signal) SleepAfter() time.Duration { return time.Second }

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.WorkflowID == "" {
		return errors.New("workflow_id is required")
	}

	// Resolve owner from the workflow if not explicitly set.
	flw, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, s.WorkflowID)
	if err != nil {
		return s.failWorkflow(ctx, errors.Wrap(err, "unable to get workflow"))
	}
	if s.OwnerID == "" {
		s.OwnerID = flw.OwnerID
	}
	if s.OwnerType == "" {
		s.OwnerType = flw.OwnerType
	}

	// Resolve queue names from owner type if not explicitly set.
	if s.StepGroupQueueName == "" || s.StepQueueName == "" || s.StepTargetQueueName == "" {
		switch s.OwnerType {
		case "installs":
			if s.StepGroupQueueName == "" {
				s.StepGroupQueueName = "install-workflow-step-groups"
			}
			if s.StepQueueName == "" {
				s.StepQueueName = "install-workflow-steps"
			}
			if s.StepTargetQueueName == "" {
				s.StepTargetQueueName = "install-signals"
			}
		case "apps":
			if s.StepGroupQueueName == "" {
				s.StepGroupQueueName = "app-workflow-step-groups"
			}
			if s.StepQueueName == "" {
				s.StepQueueName = "app-workflow-steps"
			}
			if s.StepTargetQueueName == "" {
				s.StepTargetQueueName = "app-signals"
			}
		case "app_branches":
			if s.StepGroupQueueName == "" {
				s.StepGroupQueueName = "app-branch-workflow-step-groups"
			}
			if s.StepQueueName == "" {
				s.StepQueueName = "app-branch-workflow-steps"
			}
			if s.StepTargetQueueName == "" {
				s.StepTargetQueueName = "app-branch-signals"
			}
		default:
			return s.failWorkflow(ctx, errors.Errorf("unable to resolve queue names for owner type %s", s.OwnerType))
		}
	}

	return nil
}

// failWorkflow marks the workflow as errored and returns the error.
func (s *Signal) failWorkflow(ctx workflow.Context, err error) error {
	if s.WorkflowID != "" {
		_ = statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: s.WorkflowID,
			Status: app.CompositeStatus{
				Status:                 app.StatusError,
				StatusHumanDescription: "validation failed",
				Metadata: map[string]any{
					"error_message": err.Error(),
				},
			},
		})
	}
	return err
}

func (s *Signal) Execute(ctx workflow.Context) error {
	return s.executeFlow(ctx)
}

func (s *Signal) RegisterUpdateHandlers(ctx workflow.Context) error {
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "retry-step",
		s.retryStepHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "approve-step",
		s.approveStepHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "is-retryable",
		s.isRetryableHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "skip-step",
		s.skipStepHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "cancel-step",
		s.cancelStepHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "cancel-group",
		s.cancelGroupHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "cancel-workflow",
		s.cancelWorkflowHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "poll-next-step",
		s.pollNextStepHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "retry-group",
		s.retryGroupHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "pause-workflow",
		s.pauseWorkflowHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	return workflow.SetUpdateHandlerWithOptions(ctx, "unpause-workflow",
		s.unpauseWorkflowHandler, workflow.UpdateHandlerOptions{})
}
