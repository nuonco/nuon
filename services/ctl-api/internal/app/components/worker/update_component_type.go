package worker

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	"go.temporal.io/sdk/workflow"
)

// @temporal-gen workflow
func (w *Workflows) UpdateComponentType(ctx workflow.Context, sreq signals.RequestSignal) error {
	return activities.AwaitUpdateComponentType(ctx, activities.UpdateComponentTypeRequest{
		ComponentID: sreq.ID,
		Type:        sreq.ComponentType,
	})
}
