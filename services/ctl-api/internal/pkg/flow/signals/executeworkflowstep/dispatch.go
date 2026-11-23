package executeworkflowstep

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func EnqueueBeforeQueued(ctx workflow.Context) bool {
	return true
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
