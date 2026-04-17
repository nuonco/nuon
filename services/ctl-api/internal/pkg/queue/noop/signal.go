package noop

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const (
	NoopSignalType signal.SignalType = "noop"
)

func init() {
	catalog.Register(NoopSignalType, func() signal.Signal {
		return &NoopSignal{}
	})
}

type NoopSignal struct{}

var _ signal.Signal = (*NoopSignal)(nil)

func (e *NoopSignal) Validate(ctx workflow.Context) error {
	return nil
}

func (e *NoopSignal) Execute(ctx workflow.Context) error {
	return nil
}

func (e *NoopSignal) Type() signal.SignalType {
	return NoopSignalType
}
