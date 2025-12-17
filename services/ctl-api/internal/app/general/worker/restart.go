package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/signals"
)

func (w *Workflows) restart(ctx workflow.Context, _ signals.RequestSignal) error {
	return nil
}
