package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals"
)

// @temporal-gen-v2 workflow
func (w *Workflows) Provision(ctx workflow.Context, sreq signals.RequestSignal) error {
	return nil
}

// @temporal-gen-v2 workflow
func (w *Workflows) Created(ctx workflow.Context, sreq signals.RequestSignal) error {
	w.updateStatus(ctx, sreq.ID, app.ComponentStatusActive, "component is active")
	return nil
}
