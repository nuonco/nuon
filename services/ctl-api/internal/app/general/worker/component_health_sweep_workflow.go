package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

// ComponentHealthSweep marks components unknown once their runner goes quiet.
//
// One workflow for the whole fleet, not one per install: live installs get
// their verdicts when a report arrives, so the only thing left to schedule is
// noticing silence.
func (w *Workflows) ComponentHealthSweep(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	resp, err := activities.AwaitSweepStaleComponentHealth(ctx, activities.SweepStaleComponentHealthRequest{})
	if err != nil {
		return err
	}

	if resp.Stale > 0 {
		l.Info("component health staleness sweep",
			zap.Int("stale", resp.Stale),
			zap.Int("enqueued", resp.Enqueued),
			zap.Bool("capped", resp.Capped),
		)
	}

	return nil
}
