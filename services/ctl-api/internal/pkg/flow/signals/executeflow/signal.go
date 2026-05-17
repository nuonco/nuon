package executeflow

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	qsignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

const SignalType qsignal.SignalType = "execute-workflow"

type Signal struct {
	// WorkflowID is the ID of the workflow to execute.
	WorkflowID string `json:"workflow_id"`

	// WorkflowType / OrgID / OrgName are resolved during Validate() from the
	// in-scope workflow record so the lifecycle hook can emit workflow.lifecycle
	// events without a fresh DB lookup. Empty until Validate runs.
	WorkflowType string `json:"workflow_type,omitempty"`
	OrgID        string `json:"org_id,omitempty"`
	OrgName      string `json:"org_name,omitempty"`

	// Conductor configuration — set by the creator when enqueuing.
	StepGroupQueueName     string `json:"step_group_queue_name"`
	StepQueueName          string `json:"step_queue_name"`
	StepTargetQueueName    string `json:"step_target_queue_name"`
	GenerateStepsQueueName string `json:"generate_steps_queue_name"`
	OwnerID                string `json:"owner_id"`
	OwnerType              string `json:"owner_type"`
	// OwnerName is the human-readable owner label resolved during Validate()
	// (e.g. install/app/app_branch name). Stamped onto SignalLifecycleContext
	// so workflow lifecycle webhook payloads carry owner_name without a
	// per-event DB lookup.
	OwnerName string `json:"owner_name,omitempty"`

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

var (
	_ qsignal.Signal                     = (*Signal)(nil)
	_ qsignal.SignalWithCancel           = (*Signal)(nil)
	_ qsignal.SignalWithUpdateHandlers   = (*Signal)(nil)
	_ qsignal.SignalWithLifecycleContext = (*Signal)(nil)
)

// Cancel is invoked by the queue handler when the signal is cancelled
// externally (e.g. via clear-queue). It marks the underlying workflow object as
// cancelled so the install workflow doesn't stay in a stale in-progress state.
func (s *Signal) Cancel(ctx workflow.Context) error {
	cancelCtx, cancel := workflow.NewDisconnectedContext(ctx)
	defer cancel()

	s.cancelRequested = true

	// Best-effort cancel the active group signal.
	if s.activeGroupQueueSignalID != "" {
		client.AwaitCancelSignal(cancelCtx, s.activeGroupQueueSignalID)
	}

	// Mark the workflow as finished + cancelled.
	_ = workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowFinishedAtByID(cancelCtx, s.WorkflowID)
	_ = statusactivities.AwaitPkgStatusUpdateFlowStatus(cancelCtx, statusactivities.UpdateStatusRequest{
		ID: s.WorkflowID,
		Status: app.CompositeStatus{
			Status:                 app.StatusCancelled,
			StatusHumanDescription: "workflow cancelled",
		},
	})

	return nil
}

func (s *Signal) Type() qsignal.SignalType  { return SignalType }
func (s *Signal) SleepAfter() time.Duration { return time.Second }

// LifecycleContext exposes the workflow identity + owner so lifecycle hooks
// can emit workflow.lifecycle.* webhook events without leaking inner-signal
// taxonomy. Workflow type and org id are stamped during Validate from the
// in-scope workflow record.
func (s *Signal) LifecycleContext() qsignal.SignalLifecycleContext {
	return qsignal.SignalLifecycleContext{
		OrgID:        s.OrgID,
		OrgName:      s.OrgName,
		WorkflowID:   s.WorkflowID,
		WorkflowType: s.WorkflowType,
		OwnerID:      s.OwnerID,
		OwnerType:    s.OwnerType,
		OwnerName:    s.OwnerName,
	}
}

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
	if s.WorkflowType == "" {
		s.WorkflowType = string(flw.Type)
	}
	if s.OrgID == "" {
		s.OrgID = flw.OrgID
	}
	// OrgName is preloaded by GetFlow (id + name only) so this stamping is
	// free — no extra query at validate time, no query at webhook emit time.
	if s.OrgName == "" {
		s.OrgName = flw.Org.Name
	}
	// OwnerName is resolved by GetFlow via a single PK lookup against the
	// matching polymorphic owner table (installs/apps/app_branches). Stamping
	// it here removes the per-event lookupInstallName query in the webhook hook.
	if s.OwnerName == "" {
		s.OwnerName = flw.OwnerName
	}

	// Resolve queue names from owner type if not explicitly set.
	if s.StepGroupQueueName == "" || s.StepQueueName == "" || s.StepTargetQueueName == "" || s.GenerateStepsQueueName == "" {
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
			if s.GenerateStepsQueueName == "" {
				s.GenerateStepsQueueName = "install-generate-steps"
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
			if s.GenerateStepsQueueName == "" {
				s.GenerateStepsQueueName = "app-generate-steps"
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
			if s.GenerateStepsQueueName == "" {
				s.GenerateStepsQueueName = "app-branch-generate-steps"
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
