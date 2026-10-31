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

	// Resident keeps the workflow alive after it runs 0->end: instead of
	// completing, the execute loop parks to accept run-a-step-in-between
	// updates (append/retry a step). Gated so ordinary workflows keep exact
	// run-to-completion semantics. Bounded by residentIdleTimeout — on idle
	// the loop returns cleanly and the workflow re-warms on the next dispatch.
	Resident bool `json:"resident,omitempty"`

	ResidentIdleTimeout time.Duration `json:"resident_idle_timeout,omitempty"`

	// Resume state — set by update handlers (approve/retry/skip) to wake the
	// main execute loop when it is waiting after an approval pause or error.
	resumeRequested bool
	resumeRunType   app.WorkflowRunType
	resumeStepID    string
	resumeStartIdx  int

	// appendRequested is set by the append-step update handler to wake a
	// parked resident workflow and run a freshly-added step.
	appendRequested bool

	// hostFinished is set the instant Execute returns. A resident host stays
	// open briefly afterwards (completion callbacks, handler drain), and a
	// wake-flag update accepted in that window would persist its step rows but
	// never run them. Validators reject such updates so the client retries
	// against a fresh run, which re-warms.
	hostFinished bool

	// updatesInFlight counts resident update handlers that
	// are currently executing. Those handlers persist their step rows before
	// they set appendRequested/resumeRequested, so a resident host that idles
	// out on its timer alone could close while a step is still being written
	// and orphan it until the next dispatch re-warms the host. parkResident
	// will not idle out while this counter is non-zero.
	updatesInFlight int
	updatesStarted  int
	// mutatingUpdatesStarted counts only updates that change flow state
	// (retry/append/skip/cancel/...). Read-only updates (poll-next-step,
	// is-retryable) are excluded so they can never re-warm a terminal host
	// into re-driving the conductor (see AutoExecuteReady).
	mutatingUpdatesStarted int
	retryInFlight          map[string]bool

	// Cancel state — set by cancel update handlers.
	cancelRequested bool

	// activeGroupQueueSignalID is the queue signal ID of the currently
	// executing group. Set by executeGroup, used by cancelWorkflowHandler
	// to actively cancel the running group.
	activeGroupQueueSignalID string

	// groupDispatchSeq scopes each group dispatch's completion signal per attempt.
	groupDispatchSeq int

	// awaitingResume: true while the main loop is parked awaiting resume/cancel.
	awaitingResume bool

	// executeStarted: true once executeFlow() has begun running. A retry-step
	// update that lands on a freshly re-warmed resident host (before the loop
	// reaches its parked state) must still clone+queue the retry, so the retry
	// handler treats "not started yet" like the parked case.
	executeStarted bool

	// Pause state — set by "pause-workflow" update handler. When true, the
	// flow will pause after the current group completes.
	pauseRequested bool

	// Metrics dependencies injected via SignalWithParams.
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

	// Mark the workflow as finished + cancelled with durable metadata.
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

// SleepAfter caches a finished non-resident Handler briefly so a follow-up
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

// AutoExecuteOnTerminalStart re-enters Execute when a resident host is
// (re)started by update-with-start after it idled out and completed. See the
// queue handler's re-warm path.
func (s *Signal) AutoExecuteOnTerminalStart() bool { return s.Resident }

func (s *Signal) AutoExecuteReady() bool { return s.mutatingUpdatesStarted > 0 }

// AutoExecuteDeclined reports that this re-warm was triggered only by
// read-only updates: at least one update ran, none of them were mutating, and
// none are still in flight. The Handler finishes instead of re-driving the
// conductor on a terminal flow.
func (s *Signal) AutoExecuteDeclined() bool {
	return s.updatesStarted > 0 && s.mutatingUpdatesStarted == 0 && s.updatesInFlight == 0
}

func (s *Signal) beginUpdate() func() {
	s.updatesStarted++
	s.mutatingUpdatesStarted++
	s.updatesInFlight++
	return func() { s.updatesInFlight-- }
}

// beginReadOnlyUpdate tracks an update that observes flow state without
// changing it. It still holds updatesInFlight (so parkResident cannot idle
// out mid-read) but does not arm the auto-rewarm execute path.
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
	ctx = cctx.SetWorkflowTelemetryWorkflowContext(ctx, s.workflowTelemetry())
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
	ctx = cctx.SetWorkflowTelemetryWorkflowContext(ctx, s.workflowTelemetry())

	// Resolve queue names from owner type if not explicitly set.
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
	ctx = cctx.SetWorkflowTelemetryWorkflowContext(ctx, s.workflowTelemetry())

	s.hostFinished = false
	defer func() { s.hostFinished = true }()
	return s.executeFlow(ctx)
}

// HostFinishedRejection is the update-rejection message a resident host
// returns for a wake-flag update that arrives after its conductor exited. The
// flow client matches on it to retry against a fresh Handler run.
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
