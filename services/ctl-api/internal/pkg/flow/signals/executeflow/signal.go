package executeflow

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/metrics"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	qsignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

const SignalType qsignal.SignalType = "execute-workflow"

type Signal struct {
	WorkflowID string `json:"workflow_id"`

	WorkflowType string `json:"workflow_type,omitempty"`
	OrgID        string `json:"org_id,omitempty"`
	OrgName      string `json:"org_name,omitempty"`

	StepGroupQueueName     string `json:"step_group_queue_name"`
	StepQueueName          string `json:"step_queue_name"`
	StepTargetQueueName    string `json:"step_target_queue_name"`
	GenerateStepsQueueName string `json:"generate_steps_queue_name"`
	OwnerID                string `json:"owner_id"`
	OwnerType              string `json:"owner_type"`
	OwnerName              string `json:"owner_name,omitempty"`

	Resident bool `json:"resident,omitempty"`

	ResidentIdleTimeout time.Duration `json:"resident_idle_timeout,omitempty"`

	resumeRequested bool
	resumeRunType   app.WorkflowRunType
	resumeStepID    string
	resumeStartIdx  int

	appendRequested bool

	hostFinished bool

	updatesInFlight        int
	updatesStarted         int
	mutatingUpdatesStarted int
	retryInFlight          map[string]bool

	cancelRequested bool

	activeGroupQueueSignalID string

	groupDispatchSeq int

	awaitingResume bool

	executeStarted bool

	pauseRequested bool

	mw  metrics.Writer
	v   *validator.Validate
	tmw tmetrics.Writer
}

func NewSignal(workflowID string) *Signal {
	return &Signal{
		WorkflowID: workflowID,
		Resident:   true,
	}
}

var (
	_ qsignal.Signal                      = (*Signal)(nil)
	_ qsignal.SignalWithCancel            = (*Signal)(nil)
	_ qsignal.SignalWithUpdateHandlers    = (*Signal)(nil)
	_ qsignal.SignalWithLifecycleContext  = (*Signal)(nil)
	_ qsignal.SignalWithParams            = (*Signal)(nil)
	_ qsignal.AutoExecuteOnTerminalStart  = (*Signal)(nil)
	_ qsignal.CompletionCallbacksWorkflow = (*Signal)(nil)
)

func (s *Signal) WithParams(p *qsignal.Params) {
	s.mw = p.MW
	s.v = p.V
}

func (s *Signal) Cancel(ctx workflow.Context) error {
	cancelCtx, cancel := workflow.NewDisconnectedContext(ctx)
	defer cancel()

	s.cancelRequested = true

	if s.activeGroupQueueSignalID != "" {
		client.AwaitCancelSignal(cancelCtx, s.activeGroupQueueSignalID)
	}

	_ = workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowFinishedAtByID(cancelCtx, s.WorkflowID)
	_ = statusactivities.AwaitPkgStatusUpdateFlowStatus(cancelCtx, statusactivities.UpdateStatusRequest{
		ID: s.WorkflowID,
		Status: app.CompositeStatus{
			Status:                 app.StatusCancelled,
			StatusHumanDescription: "workflow cancelled",
			Metadata: map[string]any{
				"cancel_requested_at": workflow.Now(cancelCtx).Unix(),
			},
		},
	})

	return nil
}

func (s *Signal) Type() qsignal.SignalType { return SignalType }

// why: SleepAfter caches a finished non-resident Handler briefly so a follow-up
// signal can reuse it via update-with-start. A resident host returns 0: once its
// Execute() returns the conductor loop is gone, so keeping the Handler alive
// would let an append-step/retry-step land on a finished run that can never
// consume it — the re-warm path (AutoExecuteOnTerminalStart) handles reuse
// instead.
func (s *Signal) SleepAfter() time.Duration {
	if s.Resident {
		return 0
	}
	return time.Second
}

func (s *Signal) AutoExecuteOnTerminalStart() bool { return s.Resident }

func (s *Signal) AutoExecuteReady() bool { return s.mutatingUpdatesStarted > 0 }

func (s *Signal) AutoExecuteDeclined() bool {
	return s.updatesStarted > 0 && s.mutatingUpdatesStarted == 0 && s.updatesInFlight == 0
}

func (s *Signal) beginUpdate() func() {
	s.updatesStarted++
	s.mutatingUpdatesStarted++
	s.updatesInFlight++
	return func() { s.updatesInFlight-- }
}

func (s *Signal) beginReadOnlyUpdate() func() {
	s.updatesStarted++
	s.updatesInFlight++
	return func() { s.updatesInFlight-- }
}

func (s *Signal) CompletionCallbacksWorkflowID() string {
	if !s.Resident {
		return ""
	}
	return s.WorkflowID
}

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

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.WorkflowID == "" {
		return errors.New("workflow_id is required")
	}

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
	ctx = cctx.SetWorkflowTelemetryWorkflowContext(ctx, s.workflowTelemetry())
	if s.OrgID == "" {
		s.OrgID = flw.OrgID
	}
	if s.OrgName == "" {
		s.OrgName = flw.Org.Name
	}
	if s.OwnerName == "" {
		s.OwnerName = flw.OwnerName
	}
	ctx = cctx.SetWorkflowTelemetryWorkflowContext(ctx, s.workflowTelemetry())

	if s.StepGroupQueueName == "" || s.StepQueueName == "" || s.StepTargetQueueName == "" || s.GenerateStepsQueueName == "" {
		spec, ok := queuenames.Flow(s.OwnerType)
		if !ok {
			return s.failWorkflow(ctx, errors.Errorf("unable to resolve queue names for owner type %s", s.OwnerType))
		}
		if s.StepGroupQueueName == "" {
			s.StepGroupQueueName = spec.StepGroups
		}
		if s.StepQueueName == "" {
			s.StepQueueName = spec.Steps
		}
		if s.StepTargetQueueName == "" {
			s.StepTargetQueueName = spec.StepTargets
		}
		if s.GenerateStepsQueueName == "" {
			s.GenerateStepsQueueName = spec.GenerateSteps
		}
	}

	return nil
}

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
	ctx = cctx.SetWorkflowTelemetryWorkflowContext(ctx, s.workflowTelemetry())

	s.hostFinished = false
	defer func() { s.hostFinished = true }()
	return s.executeFlow(ctx)
}

const HostFinishedRejection = "resident host finished; retry against a fresh run"

func (s *Signal) rejectIfHostFinished() error {
	if s.Resident && s.hostFinished {
		return errors.New(HostFinishedRejection)
	}
	return nil
}

func liveHostValidator[T any](s *Signal) func(workflow.Context, T) error {
	return func(workflow.Context, T) error { return s.rejectIfHostFinished() }
}

func (s *Signal) RegisterUpdateHandlers(ctx workflow.Context) error {
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "retry-step",
		s.retryStepHandler, workflow.UpdateHandlerOptions{Validator: liveHostValidator[RetryStepRequest](s)}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "approve-step",
		s.approveStepHandler, workflow.UpdateHandlerOptions{Validator: liveHostValidator[ApproveStepRequest](s)}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "is-retryable",
		s.isRetryableHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "skip-step",
		s.skipStepHandler, workflow.UpdateHandlerOptions{Validator: liveHostValidator[SkipStepRequest](s)}); err != nil {
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
		s.retryGroupHandler, workflow.UpdateHandlerOptions{Validator: liveHostValidator[RetryGroupRequest](s)}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "pause-workflow",
		s.pauseWorkflowHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "unpause-workflow",
		s.unpauseWorkflowHandler, workflow.UpdateHandlerOptions{Validator: func(workflow.Context) error { return s.rejectIfHostFinished() }}); err != nil {
		return err
	}
	return workflow.SetUpdateHandlerWithOptions(ctx, "append-step",
		s.appendStepHandler, workflow.UpdateHandlerOptions{Validator: liveHostValidator[AppendStepRequest](s)})
}
