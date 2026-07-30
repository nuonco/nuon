package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

const (
	cleanupTriggerEventBatchSize  = 5000
	cleanupTriggerEventMaxBatches = 500
)

func (w *Workflows) CleanupTriggerEvents(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	continueCleanup := false
	for _, status := range []app.EventRoutingStatus{app.EventRoutingStatusRejected, app.EventRoutingStatusIgnored} {
		var total int64
		for i := 0; i < cleanupTriggerEventMaxBatches; i++ {
			resp, err := activities.AwaitDeleteOldTriggerEvents(ctx, activities.DeleteOldTriggerEventsRequest{
				BatchSize: cleanupTriggerEventBatchSize, RoutingStatus: status,
			})
			if err != nil {
				return errors.Wrapf(err, "unable to delete old %s trigger events", status)
			}
			total += resp.Deleted
			if resp.Deleted < cleanupTriggerEventBatchSize {
				break
			}
			if cleanupTriggerEventsNeedsContinuation(i+1, resp.Deleted) {
				continueCleanup = true
			}
		}
		l.Info("cleaned up old trigger events", zap.String("routing_status", string(status)), zap.Int64("deleted", total))
	}
	var scrubbed int64
	for i := 0; i < cleanupTriggerEventMaxBatches; i++ {
		resp, err := activities.AwaitScrubInactiveTriggerSecrets(ctx, activities.ScrubInactiveTriggerSecretsRequest{BatchSize: cleanupTriggerEventBatchSize})
		if err != nil {
			return errors.Wrap(err, "unable to scrub inactive trigger secrets")
		}
		scrubbed += resp.Scrubbed
		if resp.Scrubbed < cleanupTriggerEventBatchSize {
			break
		}
		if cleanupTriggerEventsNeedsContinuation(i+1, resp.Scrubbed) {
			continueCleanup = true
		}
	}
	l.Info("scrubbed inactive trigger secrets", zap.Int64("scrubbed", scrubbed))
	if continueCleanup {
		l.Info("continuing trigger event cleanup in new execution")
		return workflow.NewContinueAsNewError(ctx, w.CleanupTriggerEvents)
	}
	return nil
}

func cleanupTriggerEventsNeedsContinuation(batches int, lastDeleted int64) bool {
	return batches >= cleanupTriggerEventMaxBatches && lastDeleted >= cleanupTriggerEventBatchSize
}
