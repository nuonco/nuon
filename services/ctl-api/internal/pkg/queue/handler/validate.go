package handler

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
)

const ValidateUpdateName string = "validate"

const validateUpdateType = handlerTypeUpdate

type ValidateResponse struct{}

func (h *handler) validateHandler(ctx workflow.Context) (*ValidateResponse, error) {
	if h.sig == nil {
		return nil, errors.New("signal was empty can not proceed")
	}

	// mark the signal as in-progress in the DB
	_ = activities.AwaitUpdateQueueSignalStatus(ctx, &activities.UpdateQueueSignalStatusRequest{
		QueueSignalID: h.queueSignalID,
		Status:        app.StatusInProgress,
	})

	if err := h.sig.Validate(ctx); err != nil {
		return nil, errors.Wrap(err, "validate method failed")
	}

	return nil, nil
}
