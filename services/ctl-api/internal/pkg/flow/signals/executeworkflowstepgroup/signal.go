package executeworkflowstepgroup

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

const SignalType signal.SignalType = "execute-workflow-step-group"

// Directive constants — duplicated to avoid import cycle with the flow package.
const (
	DirectiveContinue      = "continue"
	DirectiveStop          = "stop"
	DirectiveRetry         = "retry"
	DirectiveRetryGroup    = "retry-group"
	DirectiveSkipGroup     = "skip-group"
	DirectiveAwaitApproval = "await-approval"
)

// Signal encapsulates the lifecycle of executing all steps within a single
// workflow step group (GroupIdx). It orchestrates sequential or parallel step
// dispatch and communicates the group outcome back to the flow signal via the
// workflow's ResultDirective field.
type Signal struct {
	WorkflowID      string `json:"workflow_id"`
	StepGroupID     string `json:"step_group_id"`
	GroupIdx        int    `json:"group_idx"`
	OwnerID         string `json:"owner_id"`
	OwnerType       string `json:"owner_type"`
	QueueName       string `json:"queue_name"`
	TargetQueueName string `json:"target_queue_name"`
	Parallel        bool   `json:"parallel"`

	// WorkflowType identifies the kind of workflow that owns this group. Set
	// at dispatch time from the in-scope *app.Workflow and forwarded to each
	// child execute-workflow-step signal so the workflow_step lifecycle
	// hook can suppress events for envelope workflows like drift_run /
	// drift_run_reprovision_sandbox without a DB lookup.
	WorkflowType string `json:"workflow_type,omitempty"`

	// OrgID / OrgName / OwnerName are pass-through fields stamped by the
	// parent execute-workflow signal. They are forwarded to the step signal
	// at dispatch time so workflow_step lifecycle webhook payloads carry
	// human-readable names without a per-event DB lookup.
	OrgID     string `json:"org_id,omitempty"`
	OrgName   string `json:"org_name,omitempty"`
	OwnerName string `json:"owner_name,omitempty"`

	// finished is set when Execute() completes (success or error). The
	// group-finished update handler blocks until this is true, then returns
	// the group's final directive.
	finished bool

	// cancelRequested is set by the cancel-group update handler.
	cancelRequested bool

	// awaitingUserAction is true when the group is blocked waiting for a
	// retry-step, skip-step, or cancel after a step failure.
	awaitingUserAction bool

	// userActionReceived is set by retry-step or skip-step update handlers
	// to wake the awaitUserAction loop.
	userActionReceived bool

	// retryGroupRequested is set by the retry-step handler when the step's signal
	// declares RetryGroup. The sequential loop exits and the directive is propagated
	// to the flow for group-level retry.
	retryGroupRequested bool

	// DerivedTimeout is set at dispatch time from the group's TimeoutSeconds.
	// When non-zero, Timeout() returns this instead of the hardcoded fallback.
	DerivedTimeout time.Duration `json:"derived_timeout,omitempty"`

	// stepSignalIDs tracks in-flight step signal IDs for cancellation propagation.
	stepSignalIDs []string

	// lastDirective tracks the last directive written to the workflow.
	// Used by Execute() to determine the correct group status after execution.
	lastDirective string
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithCancel = (*Signal)(nil)
var _ signal.SignalWithUpdateHandlers = (*Signal)(nil)
var _ signal.SignalWithTimeout = (*Signal)(nil)

func (s *Signal) Timeout() time.Duration {
	if s.DerivedTimeout > 0 {
		return s.DerivedTimeout
	}
	return 2 * time.Hour
}

func (s *Signal) Type() signal.SignalType   { return SignalType }
func (s *Signal) SleepAfter() time.Duration { return time.Second }

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.WorkflowID == "" {
		return errors.New("workflow_id is required")
	}
	if s.OwnerID == "" {
		return errors.New("owner_id is required")
	}
	if s.OwnerType == "" {
		return errors.New("owner_type is required")
	}
	// QueueName and TargetQueueName are optional when all steps in the group
	// specify their own SignalQueueOwnerID for per-step routing.
	return nil
}

// RegisterUpdateHandlers registers group-level update handlers.
func (s *Signal) RegisterUpdateHandlers(ctx workflow.Context) error {
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "cancel-group",
		s.cancelGroupHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "retry-step",
		s.retryStepHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "cancel-step",
		s.cancelStepHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "approve-step",
		s.approveStepHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "skip-step",
		s.skipStepHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "group-finished",
		s.groupFinishedHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	return nil
}

func (s *Signal) cancelGroupHandler(ctx workflow.Context) error {
	s.cancelRequested = true
	return nil
}

// Cancel propagates cancellation to all in-flight step signals.
func (s *Signal) Cancel(ctx workflow.Context) error {
	cancelCtx, cancel := workflow.NewDisconnectedContext(ctx)
	defer cancel()

	l, _ := log.WorkflowLogger(cancelCtx)

	for _, qsID := range s.stepSignalIDs {
		if _, err := client.AwaitCancelSignal(cancelCtx, qsID); err != nil {
			if l != nil {
				l.Warn("failed to cancel step signal",
					zap.String("queue_signal_id", qsID),
					zap.Error(err))
			}
		}
	}

	// Mark all incomplete steps in this group as cancelled
	steps, err := s.getGroupSteps(cancelCtx)
	if err != nil {
		return err
	}

	for _, step := range steps {
		if isTerminalStatus(step.Status.Status) {
			continue
		}
		statusactivities.AwaitPkgStatusUpdateFlowStepStatus(cancelCtx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status: app.StatusCancelled,
			},
		})
	}

	// Mark the group itself as cancelled
	s.updateGroupStatus(cancelCtx, app.CompositeStatus{
		Status:                 app.StatusCancelled,
		StatusHumanDescription: "group cancelled",
	})

	return nil
}

// getGroupSteps fetches all steps for this group from the database.
// When StepGroupID is set, steps are filtered by WorkflowStepGroupID;
// otherwise falls back to GroupIdx filtering for backward compatibility.
func (s *Signal) getGroupSteps(ctx workflow.Context) ([]app.WorkflowStep, error) {
	allSteps, err := activities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, activities.GetFlowStepsRequest{
		FlowID: s.WorkflowID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get flow steps")
	}

	var groupSteps []app.WorkflowStep
	for _, step := range allSteps {
		if s.StepGroupID != "" {
			if step.WorkflowStepGroupID == s.StepGroupID {
				groupSteps = append(groupSteps, step)
			}
		} else {
			if step.GroupIdx == s.GroupIdx {
				groupSteps = append(groupSteps, step)
			}
		}
	}
	return groupSteps, nil
}

// writeStepGroupDirective writes the group's result directive to the step group's
// own ResultDirective field when a StepGroupID is set. Falls back to writing to
// the workflow's ResultDirective for backward compatibility with synthetic groups.
func (s *Signal) writeStepGroupDirective(ctx workflow.Context, directive string) error {
	s.lastDirective = directive
	if s.StepGroupID != "" {
		return activities.AwaitPkgWorkflowsFlowUpdateFlowStepGroupResultDirective(ctx, activities.UpdateFlowStepGroupResultDirectiveRequest{
			StepGroupID: s.StepGroupID,
			Directive:   directive,
		})
	}
	return s.writeWorkflowDirective(ctx, directive)
}

// writeWorkflowDirective writes the group's result directive to the workflow's
// ResultDirective field so the flow signal can read it.
func (s *Signal) writeWorkflowDirective(ctx workflow.Context, directive string) error {
	return activities.AwaitPkgWorkflowsFlowUpdateFlowResultDirective(ctx, activities.UpdateFlowResultDirectiveRequest{
		FlowID:    s.WorkflowID,
		Directive: directive,
	})
}

func isTerminalStatus(status app.Status) bool {
	switch status {
	case app.StatusSuccess, app.StatusAutoSkipped, app.StatusUserSkipped,
		app.StatusDiscarded, app.StatusCancelled, app.StatusError,
		app.WorkflowStepApprovalStatusApproved,
		app.WorkflowStepNoDrift, app.WorkflowStepDrifted:
		return true
	}
	return false
}
