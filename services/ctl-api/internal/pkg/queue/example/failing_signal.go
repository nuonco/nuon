package example

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const (
	FailingSignalType signal.SignalType = "failing-signal"
)

func init() {
	catalog.Register(FailingSignalType, func() signal.Signal {
		return &FailingSignal{}
	})
}

// FailingSignal is a test signal whose Execute always returns an error.
type FailingSignal struct {
	Reason string `json:"reason"`
}

var _ signal.Signal = (*FailingSignal)(nil)

func (f *FailingSignal) Validate(ctx workflow.Context) error {
	return nil
}

func (f *FailingSignal) Execute(ctx workflow.Context) error {
	return errors.Errorf("intentional failure: %s", f.Reason)
}

func (f *FailingSignal) Type() signal.SignalType {
	return FailingSignalType
}
