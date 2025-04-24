package worker

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

const (
	defaultPollTimeout time.Duration = time.Second * 10
)

// @temporal-gen workflow
// @execution-timeout 5m
// @task-timeout 3m
func (w *Workflows) PollDependencies(ctx workflow.Context, sreq signals.RequestSignal) error {
	installID := sreq.ID

	for {
		install, err := activities.AwaitGetByInstallID(ctx, installID)
		if err != nil {
			return fmt.Errorf("unable to get install and poll: %w", err)
		}

		if install.App.Status == "active" {
			return nil
		}
		workflow.Sleep(ctx, defaultPollTimeout)
	}
}
