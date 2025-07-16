package temporal

import (
	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/metrics"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type Params struct {
	fx.In

	L  *zap.Logger
	MW metrics.Writer
	V  *validator.Validate
}

type Client interface {
	// Send sends a signal to an event loop for asynchronous processing.
	//
	// If knowing when the signal is done processing is a requirement, use SendAndWait or SendAsync.
	Send(ctx workflow.Context, id string, signal eventloop.Signal)

	// SendAsync is the same as Send, but it returns a future that allows you to decide when to
	// wait for a result.
	//
	// Prefer Send for fire-and-forget use cases, as creating the future has some overhead.
	//
	// The value and error of the returned future will be the return type of the signal handler.
	// For signal handlers with only an error return type, a successful return will be nil, nil.
	SendAsync(ctx workflow.Context, id string, signal eventloop.Signal) (workflow.Future, error)

	// SendAndWait sends a signal to an event loop, then waits for the signal to be processed
	// before returning.
	//
	// It is the same as calling SendAsync and then immediately waiting for the future to complete.
	SendAndWait(ctx workflow.Context, id string, signal eventloop.Signal) error
}

var _ Client = (*evClient)(nil)

type evClient struct {
	l  *zap.Logger
	mw metrics.Writer
	v  *validator.Validate
}

func New(params Params) Client {
	return &evClient{
		l:  params.L,
		mw: params.MW,
		v:  params.V,
	}
}
