package worker

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/activities"
	"go.temporal.io/sdk/workflow"
)

const (
	defaultPollTimeout time.Duration = time.Second * 10
)

// TODO(ja): Components don't have a status field, so we can't update them if this fails.
// Not sure if that's a problem or not.
//
// @temporal-gen workflow
// @execution-timeout 5m
// @task-timeout 3m
func (w *Workflows) PollDependencies(ctx workflow.Context, sreq signals.RequestSignal) error {
	for {
		currentApp, err := activities.AwaitGetComponentAppByComponentID(ctx, sreq.ID)
		if err != nil {
			return fmt.Errorf("unable to get component app: %w", err)
		}

		if currentApp.Status == "active" {
			return nil
		}
		if currentApp.Status == "error" {
			return fmt.Errorf("app failed: %s", currentApp.Org.StatusDescription)
		}

		workflow.Sleep(ctx, defaultPollTimeout)
	}
}
