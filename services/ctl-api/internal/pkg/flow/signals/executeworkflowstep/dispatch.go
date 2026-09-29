package executeworkflowstep

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const DispatchEnqueueBeforeQueuedVersion = "dispatch-enqueue-before-queued-v1"

func EnqueueBeforeQueued(ctx workflow.Context) bool {
	return workflow.GetVersion(ctx, DispatchEnqueueBeforeQueuedVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion
}

func DispatchFailedStatus(enqueueErr error) app.CompositeStatus {
	meta := map[string]any{}
	if enqueueErr != nil {
		meta["reason"] = enqueueErr.Error()
	}
	return app.CompositeStatus{
		Status:                 app.StatusError,
		StatusHumanDescription: "execute signal was not created",
		Metadata:               meta,
	}
}
