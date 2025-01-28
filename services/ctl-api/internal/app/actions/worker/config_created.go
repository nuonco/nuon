package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/worker/activities"
	installsignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

// @temporal-gen workflow
// @execution-timeout 1m
// @task-timeout 30s
func (w *Workflows) ConfigCreated(ctx workflow.Context, sreq signals.RequestSignal) error {
	_, err := activities.AwaitGetActionWorkflowConfigByConfigID(ctx, sreq.ConfigID)
	if err != nil {
		return errors.Wrap(err, "unable to fetch action workflow config")
	}

	installIDs, err := activities.AwaitGetActionWorkflowInstallsByActionWorkflowID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get action workflow installs")
	}

	for _, installID := range installIDs {
		w.evClient.Send(ctx, installID, &installsignals.Signal{
			Type: installsignals.OperationSyncActionWorkflowTriggers,
		})
	}

	return nil
}
