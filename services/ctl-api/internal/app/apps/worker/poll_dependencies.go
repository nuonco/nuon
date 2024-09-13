package worker

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
	"go.temporal.io/sdk/workflow"
)

const (
	defaultPollTimeout time.Duration = time.Second * 10
)

// @temporal-gen workflow
// @execution-timeout 5m
// @task-timeout 3m
func (w *Workflows) PollDependencies(ctx workflow.Context, sreq signals.RequestSignal) error {
	for {
		currentApp, err := activities.AwaitGetByAppID(ctx, sreq.ID)
		if err != nil {
			w.updateStatus(ctx, sreq.ID, app.AppStatusError, "unable to get app from database")
			return fmt.Errorf("unable to get app from database: %w", err)
		}

		if currentApp.Org.Status == "active" {
			return nil
		}

		if currentApp.Org.Status == "error" {
			// TODO(sdboyer) remove transitive error status propagation
			w.updateStatus(ctx, sreq.ID, "error", "org is in error state")
		}

		workflow.Sleep(ctx, defaultPollTimeout)
	}
}
