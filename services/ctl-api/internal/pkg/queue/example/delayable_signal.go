package example

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const (
	ControllableSignalType       signal.SignalType = "controllable-signal"
	ControllableSignalUpdateName string            = "complete"
)

func init() {
	catalog.Register(ControllableSignalType, func() signal.Signal {
		return &ControllableSignal{}
	})
}

type ControllableSignal struct {
	ShouldBlock bool `json:"should_block"`

	// Serializable failure/behavior fields — set at enqueue time, survive DB round-trip.
	FailureMessage string        `json:"failure_message,omitempty"` // non-empty → Execute returns error
	PanicMessage   string        `json:"panic_message,omitempty"`   // non-empty → Execute panics
	Timeout        time.Duration `json:"timeout,omitempty"`         // non-zero → Execute sleeps this duration

	isValidated bool
	isExecuted  bool
	wasCanceled bool
	completeCh  workflow.Channel
}

var _ signal.Signal = (*ControllableSignal)(nil)

func (c *ControllableSignal) Validate(ctx workflow.Context) error {
	c.isValidated = true
	c.completeCh = workflow.NewChannel(ctx)

	if err := workflow.SetUpdateHandler(ctx, ControllableSignalUpdateName, c.completeHandler); err != nil {
		return err
	}

	return nil
}

func (c *ControllableSignal) Execute(ctx workflow.Context) error {
	if c.PanicMessage != "" {
		panic(c.PanicMessage)
	}
	if c.FailureMessage != "" {
		return errors.New(c.FailureMessage)
	}
	if c.Timeout > 0 {
		_ = workflow.Sleep(ctx, c.Timeout)
		return errors.New("signal timed out")
	}
	if c.ShouldBlock {
		c.completeCh.Receive(ctx, nil)
		if ctx.Err() != nil {
			c.wasCanceled = true
			return ctx.Err()
		}
	}

	c.isExecuted = true
	return nil
}

func (c *ControllableSignal) completeHandler(ctx workflow.Context) error {
	c.completeCh.Send(ctx, true)
	return nil
}

func (c *ControllableSignal) Type() signal.SignalType {
	return ControllableSignalType
}
