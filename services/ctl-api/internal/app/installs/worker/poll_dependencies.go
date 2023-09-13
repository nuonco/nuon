package worker

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"go.temporal.io/sdk/workflow"
)

const (
	defaultPollTimeout time.Duration = time.Second * 10
)

func (w *Workflows) pollDependencies(ctx workflow.Context, installID string) error {
	for {
		var install app.Install
		if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
			InstallID: installID,
		}, &install); err != nil {
			return fmt.Errorf("unable to get install and poll: %w", err)
		}

		if install.App.Status == "active" {
			return nil
		}
		if install.App.Status == "error" {
			return fmt.Errorf("app failed: %s", install.App.StatusDescription)
		}
		workflow.Sleep(ctx, defaultPollTimeout)
	}

	return nil
}
