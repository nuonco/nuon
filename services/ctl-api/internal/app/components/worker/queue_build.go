package worker

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) queueBuild(ctx workflow.Context, cmpID string) error {
	var cmp app.Component
	if err := w.defaultExecGetActivity(ctx, w.acts.GetComponent, activities.GetComponentRequest{
		ComponentID: cmpID,
	}, &cmp); err != nil {
		return fmt.Errorf("unable to get component: %w", err)
	}

	var cmpBuild app.ComponentBuild
	if err := w.defaultExecGetActivity(ctx, w.acts.QueueComponentBuild, activities.QueueComponentBuildRequest{
		CreatedByID: cmp.CreatedByID,
		ComponentID: cmpID,
		OrgID:       cmp.OrgID,
	}, &cmpBuild); err != nil {
		return fmt.Errorf("unable to queue component build: %w", err)
	}

	return nil
}
