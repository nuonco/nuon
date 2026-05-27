package handler

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func (h *handler) initializeState(ctx workflow.Context) error {
	queueSignal, err := activities.AwaitGetQueueSignalByQueueSignalID(ctx, h.queueSignalID)
	if err != nil {
		return errors.Wrap(err, "unable to get queue signal")
	}

	// If the signal has an expiry time and we're past it, terminate without processing.
	if queueSignal.ExpiresAt != nil && workflow.Now(ctx).After(*queueSignal.ExpiresAt) {
		_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
			QueueSignalID:     h.queueSignalID,
			Status:            app.StatusError,
			StatusDescription: "signal expired",
			Metadata: map[string]any{
				"expired_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
			},
		})
		h.setFinished(app.StatusError, "signal expired")
		return nil
	}

	// Detect abandoned previous runs by checking for started-but-not-finished timestamps.
	// If a previous handler crashed mid-validate or mid-execute, the started_at timestamp
	// will exist but the finished_at timestamp will not.
	if meta := queueSignal.Status.Metadata; meta != nil {
		// If init previously failed, the signal is already in a terminal error state.
		if _, initFailed := meta["init_failed_at"]; initFailed {
			h.setFinished(app.StatusError, "previous init failed; signal already in terminal state")
			return nil
		}

		_, valStarted := meta["validate_started_at"]
		_, valFinished := meta["validate_finished_at"]
		if valStarted && !valFinished {
			desc := "previous execution was abandoned (crashed mid-validate), marking as failed"
			_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
				QueueSignalID:     h.queueSignalID,
				Status:            app.StatusError,
				StatusDescription: desc,
				Metadata: map[string]any{
					"validate_finished_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
				},
			})
			h.setFinished(app.StatusError, desc)
			return nil
		}

		_, execStarted := meta["execute_started_at"]
		_, execFinished := meta["execute_finished_at"]
		if execStarted && !execFinished {
			desc := "previous execution was abandoned (crashed mid-execute), marking as failed"
			_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
				QueueSignalID:     h.queueSignalID,
				Status:            app.StatusError,
				StatusDescription: desc,
				Metadata: map[string]any{
					"execute_finished_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
				},
			})
			h.setFinished(app.StatusError, desc)
			return nil
		}
	}

	sig, err := activities.AwaitGetQueueSignalSignalByQueueSignalID(ctx, h.queueSignalID)
	if err != nil {
		if catalog.IsSignalTypeNotRegistered(err) {
			_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
				QueueSignalID:     h.queueSignalID,
				Status:            app.StatusPending,
				StatusDescription: "signal type not registered in current build; parked for redeploy",
			})
			return nil
		}
		return errors.Wrap(err, "unable to get signal")
	}
	h.sig = sig

	h.queueSignal = queueSignal

	signal.ApplyParams(h.sig, &signal.Params{
		Cfg:           h.cfg,
		V:             h.v,
		MW:            h.mw,
		QueueSignalID: h.queueSignalID,
	})

	if err := signal.ApplyInit(h.sig, ctx); err != nil {
		initErr := &signal.SignalErrInit{Err: err}
		_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
			QueueSignalID:     h.queueSignalID,
			Status:            app.StatusError,
			StatusDescription: initErr.Error(),
			Metadata: map[string]any{
				"init_failed_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
			},
		})
		return initErr
	}

	return nil
}
