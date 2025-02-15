package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/signals"
)

func (w *Workflows) created(ctx workflow.Context, _ signals.RequestSignal) error {
	// our logic goes here
	return nil
}
