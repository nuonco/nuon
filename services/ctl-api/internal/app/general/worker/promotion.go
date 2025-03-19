package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/signals"
)

// @temporal-gen workflow
// @execution-timeout 10m
// @task-timeout 30s
func (w *Workflows) Promotion(ctx workflow.Context, _ signals.RequestSignal) error {
	if err := w.RestartOrgEventLoops(ctx); err != nil {
		return errors.Wrap(err, "unable to restart org event loops")
	}

	if err := w.RestartOrgRunners(ctx); err != nil {
		return errors.Wrap(err, "unable to restart org runners")
	}

	return nil
}
