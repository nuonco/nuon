package worker

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

const (
	defaultPollTimeout time.Duration = time.Second * 10
)

func (w *Workflows) pollDependencies(ctx workflow.Context, installID string) error {
	w.writeInstallEvent(ctx, installID, signals.OperationPollDependencies, app.OperationStatusStarted)

	for {
		var install app.Install
		if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
			InstallID: installID,
		}, &install); err != nil {
			w.updateStatus(ctx, installID, StatusError, "unable to get install from database")
			w.writeInstallEvent(ctx, installID, signals.OperationPollDependencies, app.OperationStatusFailed)
			return fmt.Errorf("unable to get install and poll: %w", err)
		}

		if install.App.Status == "active" {
			return nil
		}
		if install.App.Status == "error" {
			w.updateStatus(ctx, installID, StatusError, "app is in error state")
		}
		workflow.Sleep(ctx, defaultPollTimeout)
	}

	w.writeInstallEvent(ctx, installID, signals.OperationPollDependencies, app.OperationStatusFinished)
	return nil
}
