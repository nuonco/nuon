package worker

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	"go.temporal.io/sdk/workflow"
)

// @temporal-gen workflow
func (w *Workflows) QueueBuild(ctx workflow.Context, sreq signals.RequestSignal) error {
	cmp, err := activities.AwaitGetComponentByComponentID(ctx, sreq.ID)
	if err != nil {
		return fmt.Errorf("unable to get component: %w", err)
	}

	_, err = activities.AwaitQueueComponentBuild(ctx, activities.QueueComponentBuildRequest{
		CreatedByID: cmp.CreatedByID,
		ComponentID: sreq.ID,
		OrgID:       cmp.OrgID,
	})
	if err != nil {
		return fmt.Errorf("unable to queue component build: %w", err)
	}

	return nil
}
