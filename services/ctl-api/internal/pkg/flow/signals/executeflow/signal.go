package executeflow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	qsignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

const SignalType qsignal.SignalType = "execute-flow"

type Signal struct {
	// WorkflowID is the ID of the workflow to execute.
	WorkflowID string `json:"workflow_id"`

	// Conductor configuration — set by the creator when enqueuing.
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

	// Cancel state — set by "cancel-step" update handler.
	cancelRequested bool
}

var _ qsignal.Signal = (*Signal)(nil)
var _ qsignal.SignalWithUpdateHandlers = (*Signal)(nil)

func (s *Signal) Type() qsignal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.WorkflowID == "" {
		return errors.New("workflow_id is required")
	}

	// Resolve owner from the workflow if not explicitly set.
	flw, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, s.WorkflowID)
	if err != nil {
		return errors.Wrap(err, "unable to get workflow")
	}
	if s.OwnerID == "" {
		s.OwnerID = flw.OwnerID
	}
	if s.OwnerType == "" {
		s.OwnerType = flw.OwnerType
	}

	// Validate owner exists.
	// NOTE: currently install-specific; will be abstracted when activities move.
	if s.OwnerType == "installs" {
		if _, err := installactivities.AwaitGetByInstallID(ctx, s.OwnerID); err != nil {
			return errors.Wrap(err, "owner not found")
		}
	}

	return nil
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
	return workflow.SetUpdateHandlerWithOptions(ctx, "poll-next-step",
		s.pollNextStepHandler, workflow.UpdateHandlerOptions{})
}
