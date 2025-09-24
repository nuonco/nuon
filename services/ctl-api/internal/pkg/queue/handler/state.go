package handler

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/activities"
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

	return nil
}
