package driftdetected

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

// Dispatch enqueues a drift-detected signal onto the install-signals queue
// and waits for it to reach a terminal phase. The signal itself is a no-op —
// it exists so the queue dispatcher emits lifecycle events that the interests
// classifier maps onto event:drift.detected, fanning out to webhook / Slack
// subscribers that opted into per-resource `drift_detected: true`.
//
// The caller is responsible for populating InstallID, InstallWorkflowID,
// WorkflowStepID, OwnerID, and OwnerType on sig before calling.
func Dispatch(ctx workflow.Context, sig *Signal) error {
	resp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         sig.InstallID,
		OwnerType:       "installs",
		QueueName:       installSignalsQueueName,
		Signal:          sig,
		SignalOwnerID:   sig.WorkflowStepID,
		SignalOwnerType: installWorkflowStepsOwnerType,
	})
	if err != nil {
		return errors.Wrap(err, "unable to enqueue drift-detected signal")
	}

	if _, err := client.AwaitQueueSignal(ctx, resp.QueueSignalID); err != nil {
		return errors.Wrap(err, "drift-detected signal failed")
	}
	return nil
}
