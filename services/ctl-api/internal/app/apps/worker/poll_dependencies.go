package worker

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
	"go.temporal.io/sdk/workflow"
)

const (
	defaultPollTimeout time.Duration = time.Second * 10
)

func (w *Workflows) pollDependencies(ctx workflow.Context, appID string) error {
	for {
		currentApp, err := activities.AwaitGetByAppID(ctx, appID)
		if err != nil {
			w.updateStatus(ctx, appID, app.AppStatusError, "unable to get app from database")
			return fmt.Errorf("unable to get app from database: %w", err)
		}

		if currentApp.Org.Status == "active" {
			return nil
		}

		if currentApp.Org.Status == "error" {
			// TODO(sdboyer) remove transitive error status propagation
			w.updateStatus(ctx, appID, "error", "org is in error state")
		}

		workflow.Sleep(ctx, defaultPollTimeout)
	}
}
