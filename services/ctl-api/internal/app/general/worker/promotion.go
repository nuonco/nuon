package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/cockroachdb/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/signals"
)

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 30s
func (w *Workflows) Promotion(ctx workflow.Context, _ signals.RequestSignal) error {
	grp := workflow.NewWaitGroup(ctx)
	grp.Add(2)

	var orgerr, runerr error
	workflow.Go(ctx, func(ctx workflow.Context) {
		if err := w.RestartOrgEventLoops(ctx); err != nil {
			orgerr = errors.Wrap(err, "unable to restart org event loops")
		}
		grp.Done()
	})

	workflow.Go(ctx, func(ctx workflow.Context) {
		if err := w.RestartOrgRunners(ctx); err != nil {
			runerr = errors.Wrap(err, "unable to restart org runners")
		}
		grp.Done()
	})

	grp.Wait(ctx)
	return errors.Join(orgerr, runerr)
}
