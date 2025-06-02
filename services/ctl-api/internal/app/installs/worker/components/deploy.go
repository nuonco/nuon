package components

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Workflows) errorResponse(ctx workflow.Context, sreq signals.RequestSignal, deployID, installComponentID, message string, err error) error {
	if sreq.ForceDelete {
		w.updateDeployStatusWithoutStatusSync(ctx, deployID, app.InstallDeployStatusInactive, message)
		w.updateInstallComponentStatus(ctx, installComponentID, app.InstallComponentStatusDeleteFailed, message)
		if err := activities.AwaitDeleteInstallComponent(ctx, activities.DeleteInstallComponentRequest{
			InstallComponentID: installComponentID,
		}); err != nil {
			return errors.Wrap(err, "unable to delete install component")
		}
		return nil
	}

	w.updateDeployStatus(ctx, deployID, app.InstallDeployStatusError, message)
	return errors.Wrap(err, message)
}
