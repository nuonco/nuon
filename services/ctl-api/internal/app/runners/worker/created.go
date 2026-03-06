package worker

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	"go.temporal.io/sdk/workflow"
)

// @temporal-gen-v2 workflow
func (w *Workflows) Created(ctx workflow.Context, sreq signals.RequestSignal) error {
	return nil
}
