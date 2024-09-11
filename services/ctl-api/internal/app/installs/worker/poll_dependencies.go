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
		install, err := activities.AwaitGetByInstallID(ctx, installID)
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
