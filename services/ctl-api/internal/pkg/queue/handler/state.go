package handler

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
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
		Cfg: h.cfg,
		V:   h.v,
	})

	return nil
}
