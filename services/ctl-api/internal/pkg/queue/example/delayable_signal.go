package example

import (
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
	signal.Hooks
	ShouldBlock bool `json:"should_block"`

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
