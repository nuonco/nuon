package handler

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func (h *handler) initializeState(ctx workflow.Context) error {
	queueSignal, err := activities.AwaitGetQueueSignalByQueueSignalID(ctx, h.queueSignalID)
	if err != nil {
		return errors.Wrap(err, "unable to get queue signal")
	}

	sig, err := activities.AwaitGetQueueSignalSignalByQueueSignalID(ctx, h.queueSignalID)
	if err != nil {
		return errors.Wrap(err, "unable to get signal")
	}
	if sig == nil {
		panic("signal was nil")
	}

	h.queueSignal = queueSignal
	h.sig = sig

	signal.ApplyParams(h.sig, &signal.Params{
		Cfg:           h.cfg,
		V:             h.v,
		QueueSignalID: h.queueSignalID,
	})

	// Populate infrastructure fields on the signal's Hooks struct.
	hooks := h.sig.GetHooks()
	hooks.QueueSignalID = h.queueSignalID
	hooks.QueueID = h.queueID
	if queueSignal != nil {
		hooks.SignalType = signal.SignalType(queueSignal.Type)
		if queueSignal.OrgID != nil {
			hooks.OrgID = *queueSignal.OrgID
		}
	}

	if err := signal.ApplyInit(h.sig, ctx); err != nil {
		initErr := &signal.SignalErrInit{Err: err}
		_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
			QueueSignalID:     h.queueSignalID,
			Status:            app.StatusError,
			StatusDescription: initErr.Error(),
		})
		return initErr
	}

	return nil
}
