package driftdetected

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

func Dispatch(ctx workflow.Context, sig *Signal) error {
	cb := callback.New(ctx, sig.WorkflowStepID)
	_, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         sig.InstallID,
		OwnerType:       "installs",
		QueueName:       queuenames.InstallSignalsQueueName,
		Signal:          sig,
		SignalOwnerID:   sig.WorkflowStepID,
		SignalOwnerType: installWorkflowStepsOwnerType,
		Callback:        cb,
	})
	if err != nil {
		return errors.Wrap(err, "unable to enqueue drift-detected signal")
	}

	if _, err := callback.AwaitWithTimeout(ctx, cb, callback.DriftDetectionTimeout); err != nil {
		return errors.Wrap(err, "drift-detected signal failed")
	}
	return nil
}
