package worker

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"go.temporal.io/sdk/workflow"
)

// @temporal-gen workflow
func (w *Workflows) Reprovision(ctx workflow.Context, sreq signals.RequestSignal) error {
	return nil
}
