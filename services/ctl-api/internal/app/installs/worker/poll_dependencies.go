package worker

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

const (
	defaultPollTimeout time.Duration = time.Second * 10
)

// @temporal-gen workflow
// @execution-timeout 5m
// @task-timeout 3m
func (w *Workflows) PollDependencies(ctx workflow.Context, sreq signals.RequestSignal) error {
	installID := sreq.ID

	w.writeInstallEvent(ctx, installID, signals.OperationPollDependencies, app.OperationStatusStarted)

	for {
		install, err := helpers.AwaitGetInstallByID(ctx, installID)
		if err != nil {
			w.writeInstallEvent(ctx, installID, signals.OperationPollDependencies, app.OperationStatusFailed)
			return fmt.Errorf("unable to get install and poll: %w", err)
		}

		if install.App.Status == "active" {
			w.writeInstallEvent(ctx, installID, signals.OperationPollDependencies, app.OperationStatusFinished)
			return nil
		}
		workflow.Sleep(ctx, defaultPollTimeout)
	}
}
