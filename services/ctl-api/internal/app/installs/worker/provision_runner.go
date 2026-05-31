package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

// @temporal-gen-v2 workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) ProvisionRunner(ctx workflow.Context, sreq signals.RequestSignal) error {
	installID := sreq.ID

	_, err := activities.AwaitGet(ctx, activities.GetRequest{
		InstallID: installID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	return nil
}
