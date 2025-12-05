package example

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/signal"
)

const (
	ExampleSignalType signal.SignalType = "example-signal"
)

func init() {
	catalog.Register(ExampleSignalType, func() signal.Signal {
		return &ExampleSignal{}
	})
}

type ExampleSignal struct {
	Arg1 string `json:"arg_1"`
	Arg2 string `json:"arg_2"`

	isValidated bool
	isExecuted  bool
}

var _ signal.Signal = (*ExampleSignal)(nil)

func (e *ExampleSignal) Validate(ctx workflow.Context) error {
	e.isValidated = true
	return nil
}

func (e *ExampleSignal) Execute(ctx workflow.Context) error {
	e.isExecuted = true
	return nil
}

func (e *ExampleSignal) Type() signal.SignalType {
	return ExampleSignalType
}
