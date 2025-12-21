package flow

import (
	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/workflow"

	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	teventloop "github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop/temporal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

type WorkflowStepGenerator func(ctx workflow.Context, uf *app.Workflow) ([]*app.WorkflowStep, error)

type WorkflowConductor[DomainSignal eventloop.Signal] struct {
	Cfg        *internal.Config
	MW         tmetrics.Writer
	V          *validator.Validate
	EVClient   teventloop.Client
	Generators map[app.WorkflowType]WorkflowStepGenerator

	// ExecFnLegacy is called to actually execute the signal handler for a step.
	//
	// TODO(sdboyer) THIS IS A TEMPORARY HACK. Dispatching should be done within
	// the conductor itself.  However, we absolutely can't do it until we allow
	// certain concurrent behaviors in event loops, as it would deadlock when we
	// signal the same event loop that's running this workflow. It'll also be a
	// bit of awkward coupling to do it without totally predictable event loop
	// workflow IDs, but that's not a hard blocker.
	ExecFnLegacy func(workflow.Context, eventloop.EventLoopRequest, DomainSignal, app.WorkflowStep) error

	// ExecFn is called to execute a queue-signal-based step. Unlike ExecFnLegacy, it does not
	// require a generic DomainSignal or an EventLoopRequest — it operates directly on the
	// QueueSignal stored on the workflow step.
	ExecFn func(workflow.Context, *signaldb.SignalData, app.WorkflowStep) error

	// StepChildWorkflow controls whether QueueSignal-based steps are executed as child workflows.
	// When true, each step is run in its own child workflow (ExecuteStep). Only applies to
	// steps where QueueSignal != nil.
	StepChildWorkflow bool

	// NOTE(sdboyer) these will be used after ExecFnLegacy is removed
	// NewRequestSignal is used by the conductor to create new request signals as needed
	// during the course of flow execution.
	// NewRequestSignal func(ReqSig, SignalType) ReqSig

	// SignalIDRouter is called by the conductor to determine the ID of the event loop to which the signal for
	// a given step should be dispatched.
	//
	// The return value should be a string that is the ID of the event loop, but omitting the 'event-loop-' prefix.
	//
	// TODO(sdboyer) routing by opaque magic strings is a code smell. this can and should be done by the conductor/framework based on object identity
	// SignalIDRouter func(ReqSig) string
}
