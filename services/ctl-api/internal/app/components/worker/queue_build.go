package worker

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) queueBuild(ctx workflow.Context, cmpID string) error {
	var cmpBuild app.ComponentBuild
	if err := w.defaultExecGetActivity(ctx, w.acts.QueueComponentBuild, activities.QueueComponentBuildRequest{
		ComponentID: cmpID,
	}, &cmpBuild); err != nil {
		return fmt.Errorf("unable to queue component build: %w", err)
	}

	return nil
}
