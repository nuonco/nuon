package flow

import (
	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/workflow"

	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

type WorkflowStepGenerator func(ctx workflow.Context, uf *app.Workflow) (*app.GenerateStepsResult, error)

// LegacyRequest holds the minimal state previously carried by eventloop.EventLoopRequest.
// It is kept so existing callers can pass an ID and sandbox-mode flag through the
// conductor without importing the deleted eventloop package.
type LegacyRequest struct {
	ID          string
	SandboxMode bool
}

type WorkflowConductor[DomainSignal any] struct {
	Cfg        *internal.Config
	MW         tmetrics.Writer
	V          *validator.Validate
	Generators map[app.WorkflowType]WorkflowStepGenerator

	// ExecFnLegacy is called to actually execute the signal handler for a step.
	//
	// TODO(sdboyer) THIS IS A TEMPORARY HACK. Dispatching should be done within
	// the conductor itself.  However, we absolutely can't do it until we allow
	// certain concurrent behaviors in event loops, as it would deadlock when we
	// signal the same event loop that's running this workflow. It'll also be a
	// bit of awkward coupling to do it without totally predictable event loop
	// workflow IDs, but that's not a hard blocker.
	ExecFnLegacy func(workflow.Context, LegacyRequest, DomainSignal, app.WorkflowStep) error

	// ExecFn is called to execute a queue-signal-based step. Unlike ExecFnLegacy, it does not
	// require a generic DomainSignal or a LegacyRequest — it operates directly on the
	// QueueSignal stored on the workflow step.
	ExecFn func(workflow.Context, *signaldb.SignalData, app.WorkflowStep) error

	// StepChildWorkflow controls whether QueueSignal-based steps are executed via the
	// execute-workflow-step signal. When true, each step is dispatched through its own
	// signal execution. Only applies to steps where QueueSignal != nil.
	StepChildWorkflow bool

	// StepQueueName is the queue where the execute-workflow-step signal itself runs
	// (e.g. "install-workflow-steps"). When StepChildWorkflow is true, each step's
	// full lifecycle is dispatched to this queue.
	StepQueueName string
	// StepTargetQueueName is the queue where the inner signal (the actual step signal)
	// gets enqueued for execution (e.g. "install-signals").
	StepTargetQueueName string
	StepOwnerID         string
	StepOwnerType       string
}
