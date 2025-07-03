package flow

import (
	"github.com/go-playground/validator/v10"
	tmetrics "github.com/powertoolsdev/mono/pkg/temporal/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	teventloop "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/temporal"
	"go.temporal.io/sdk/workflow"
)

type WorkflowStepGenerator func(ctx workflow.Context, uf *app.Workflow) ([]*app.WorkflowStep, error)

type WorkflowConductor[DomainSignal eventloop.Signal] struct {
	Cfg        *internal.Config
	MW         tmetrics.Writer
	V          *validator.Validate
	EVClient   teventloop.Client
	Generators map[app.WorkflowType]WorkflowStepGenerator

	// ExecFn is called to actually execute the signal handler for a step.
	//
	// TODO(sdboyer) THIS IS A TEMPORARY HACK. Dispatching should be done within
	// the conductor itself.  However, we absolutely can't do it until we allow
	// certain concurrent behaviors in event loops, as it would deadlock when we
	// signal the same event loop that's running this workflow. It'll also be a
	// bit of awkward coupling to do it without totally predictable event loop
	// workflow IDs, but that's not a hard blocker.
	ExecFn func(workflow.Context, eventloop.EventLoopRequest, DomainSignal, app.WorkflowStep) error

	// NOTE(sdboyer) these will be used after ExecFn is removed
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
