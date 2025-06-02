package flows

import (
	"github.com/go-playground/validator/v10"
	tmetrics "github.com/powertoolsdev/mono/pkg/temporal/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

type Params struct {
	Cfg   *internal.Config
	MW    tmetrics.Writer
	V     *validator.Validate
}

type Flows struct {
	cfg   *internal.Config
	mw    tmetrics.Writer
	v     *validator.Validate
}

// The flow generator needs to be type-specific on the metadata it requires

func NewFlows(params Params) *Flows {
	return &Flows{
		cfg: params.Cfg,
		mw:  params.MW,
		v:   params.V,
	}
}

// We want the domain-specific package itself to define 1) its flows and 2) the process for generating the steps for them. The framework should take care of the rest
//
// We can declare a type-specific thing with generators for each of the known/named flow types
// These generators would be called by the (successor to) ExecuteWorkflow, which has the information about a) which exact flow to run, and b) what the input objects (e.g. install id) are
// ExecuteWorkflow itself is pretty generic, to the point where registration of it on an event loop should be very little code. Could...even just be a generically typed option
// But at that point, it's almost weird to have it as a signal on the main event loop. All we're getting there is serialization/concurrency control of the flows. And we DON'T want that to block other things on the main event loop...necessarily
// 
// But we'll fix that later.
// When an execute flow signal comes in to the main event loop, it will go to a signal handler that then uses the generic framework by registering all the generators and calling Run, much like event loops.
// We basically skirt arg typing by going through the db - the generic stuff loads the flow, which has the metadata map bucket
