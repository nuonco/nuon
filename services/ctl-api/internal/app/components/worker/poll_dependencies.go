package worker

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/activities"
	"go.temporal.io/sdk/workflow"
)

const (
	defaultPollTimeout time.Duration = time.Second * 10
)

// TODO(ja): Components don't have a status field, so we can't update them if this fails.
// Not sure if that's a problem or not.
func (w *Workflows) pollDependencies(ctx workflow.Context, componentID string) error {
	for {
		var currentApp app.App
		if err := w.defaultExecGetActivity(ctx, w.acts.GetComponentApp, activities.GetComponentAppRequest{
			ComponentID: componentID,
		}, &currentApp); err != nil {
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
