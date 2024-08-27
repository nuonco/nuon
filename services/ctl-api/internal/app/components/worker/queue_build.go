package worker

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) queueBuild(ctx workflow.Context, cmpID string) error {
	cmp, err := activities.AwaitGetComponentByComponentID(ctx, cmpID)
	if err != nil {
		return fmt.Errorf("unable to get component: %w", err)
	}

	_, err = activities.AwaitQueueComponentBuild(ctx, activities.QueueComponentBuildRequest{
		CreatedByID: cmp.CreatedByID,
		ComponentID: cmpID,
		OrgID:       cmp.OrgID,
	})
	if err != nil {
		return fmt.Errorf("unable to queue component build: %w", err)
	}

	return nil
}
