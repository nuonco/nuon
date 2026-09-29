package executeworkflowstepgroup

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

const SignalType signal.SignalType = "execute-workflow-step-group"

const (
	DirectiveContinue      = directive.StepContinue
	DirectiveStop          = directive.StepStop
	DirectiveRetry         = directive.StepRetry
	DirectiveRetryGroup    = directive.StepRetryGroup
	DirectiveSkipGroup     = directive.StepSkipGroup
	DirectiveAwaitApproval = directive.StepAwaitApproval
	DirectiveAwaitRetry    = directive.StepAwaitRetry
)

type Signal struct {
	WorkflowID      string `json:"workflow_id"`
	StepGroupID     string `json:"step_group_id"`
	GroupIdx        int    `json:"group_idx"`
	OwnerID         string `json:"owner_id"`
	OwnerType       string `json:"owner_type"`
	QueueName       string `json:"queue_name"`
	TargetQueueName string `json:"target_queue_name"`
	Parallel        bool   `json:"parallel"`

	WorkflowType string `json:"workflow_type,omitempty"`

	OrgID     string `json:"org_id,omitempty"`
	OrgName   string `json:"org_name,omitempty"`
	OwnerName string `json:"owner_name,omitempty"`
	finished  bool

	cancelRequested bool

	DerivedTimeout time.Duration `json:"derived_timeout,omitempty"`

	ResidentFlow bool `json:"resident_flow,omitempty"`

	stepSignalIDs []string

	stepDispatchSeq int

	lastDirective string

	mw metrics.Writer
}

var (
	_ signal.Signal                     = (*Signal)(nil)
	_ signal.SignalWithCancel           = (*Signal)(nil)
	_ signal.SignalWithUpdateHandlers   = (*Signal)(nil)
	_ signal.SignalWithTimeout          = (*Signal)(nil)
	_ signal.SignalWithParams           = (*Signal)(nil)
	_ signal.SignalWithLifecycleContext = (*Signal)(nil)
)

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	return signal.SignalLifecycleContext{
		OrgID:        s.OrgID,
		OrgName:      s.OrgName,
		WorkflowID:   s.WorkflowID,
		WorkflowType: s.WorkflowType,
		OwnerID:      s.OwnerID,
		OwnerType:    s.OwnerType,
		OwnerName:    s.OwnerName,
		Operation:    "execute-workflow-step-group",
		Metadata: map[string]any{
			"step_group_id": s.StepGroupID,
			"group_idx":     s.GroupIdx,
		},
	}
}

func (s *Signal) workflowTelemetry() cctx.WorkflowTelemetry {
	telemetry := cctx.WorkflowTelemetry{
		OrgID:        s.OrgID,
		OrgName:      s.OrgName,
		WorkflowID:   s.WorkflowID,
		WorkflowType: s.WorkflowType,
		OwnerID:      s.OwnerID,
		OwnerType:    s.OwnerType,
		OwnerName:    s.OwnerName,
	}
	if s.OwnerType == plugins.TableNameOf[app.Install]() {
		telemetry.InstallID = s.OwnerID
		telemetry.InstallName = s.OwnerName
	}
	return telemetry
}

func (s *Signal) WithParams(params *signal.Params) {
	s.mw = params.MW
}

func (s *Signal) Timeout() time.Duration {
	if s.DerivedTimeout != 0 {
		return s.DerivedTimeout
	}
	return 30 * 24 * time.Hour
}

func (s *Signal) UnboundedTimeout() bool { return s.DerivedTimeout < 0 }

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
	return nil
}

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

func (s *Signal) Cancel(ctx workflow.Context) error {
	s.cancelRequested = true

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

	s.updateGroupStatus(cancelCtx, app.CompositeStatus{
		Status:                 app.StatusCancelled,
		StatusHumanDescription: "group cancelled",
	})

	return nil
}

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

func (s *Signal) writeStepGroupDirective(ctx workflow.Context, d directive.Group) error {
	s.lastDirective = string(d)
	if s.StepGroupID != "" {
		return activities.AwaitPkgWorkflowsFlowUpdateFlowStepGroupResultDirective(ctx, activities.UpdateFlowStepGroupResultDirectiveRequest{
			StepGroupID: s.StepGroupID,
			Directive:   string(d),
		})
	}
	return s.writeWorkflowDirective(ctx, string(d))
}

func (s *Signal) writeWorkflowDirective(ctx workflow.Context, d string) error {
	return activities.AwaitPkgWorkflowsFlowUpdateFlowResultDirective(ctx, activities.UpdateFlowResultDirectiveRequest{
		FlowID:    s.WorkflowID,
		Directive: d,
	})
}

func (s *Signal) isWorkflowCancelled(ctx workflow.Context) bool {
	flw, err := activities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, s.WorkflowID)
	if err != nil {
		return false
	}
	if flw.Status.Status == app.StatusCancelled {
		return true
	}
	if flw.Status.Metadata != nil {
		if _, ok := flw.Status.Metadata["cancel_requested_at"]; ok {
			return true
		}
	}
	return false
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
