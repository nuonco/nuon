package example

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const (
	PanickingSignalType signal.SignalType = "panicking-signal"
)

func init() {
	catalog.Register(PanickingSignalType, func() signal.Signal {
		return &PanickingSignal{}
	})
}

// PanickingSignal is a test signal whose Execute always panics.
type PanickingSignal struct {
	signal.Hooks
	Message string `json:"message"`
}

var _ signal.Signal = (*PanickingSignal)(nil)

func (p *PanickingSignal) Validate(ctx workflow.Context) error {
	return nil
}

func (p *PanickingSignal) Execute(ctx workflow.Context) error {
	panic(p.Message)
}

func (p *PanickingSignal) Type() signal.SignalType {
	return PanickingSignalType
}
