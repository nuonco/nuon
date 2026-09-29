package example

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const (
	AwaitCallbackSignalType signal.SignalType = "await-callback-signal"
)

func init() {
	catalog.Register(AwaitCallbackSignalType, func() signal.Signal {
		return &AwaitCallbackSignal{}
	})
}

type AwaitCallbackSignal struct {
	AwaitID string `json:"await_id"`
}

var _ signal.Signal = (*AwaitCallbackSignal)(nil)

func (s *AwaitCallbackSignal) Validate(ctx workflow.Context) error {
	return nil
}

func (s *AwaitCallbackSignal) Execute(ctx workflow.Context) error {
	ref := callback.New(ctx, s.AwaitID)
	_, err := callback.AwaitWithTimeout(ctx, ref, 5*time.Minute)
	return err
}

func (s *AwaitCallbackSignal) Type() signal.SignalType {
	return AwaitCallbackSignalType
}
