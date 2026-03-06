package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type EnqueueSignalToOwnerRequest struct {
	OwnerID   string        `json:"owner_id" validate:"required"`
	OwnerType string        `json:"owner_type" validate:"required"`
	Signal    signal.Signal `json:"signal" validate:"required"`
}

type EnqueueSignalToOwnerResponse struct {
	QueueSignalID string `json:"queue_signal_id"`
	WorkflowID    string `json:"workflow_id"`
}

// EnqueueSignalToOwner sends a signal to a queue owned by a specific entity (e.g., runner, install).
// This enables cross-namespace signal sending where one namespace can trigger work in another.
//
// @temporal-gen-v2 activity
func (a *Activities) EnqueueSignalToOwner(ctx context.Context, req *EnqueueSignalToOwnerRequest) (*EnqueueSignalToOwnerResponse, error) {
	if err := a.v.Struct(req); err != nil {
		return nil, errors.Wrap(err, "invalid request")
	}

	// Find the queue by owner
	queue, err := a.queueClient.GetQueueByOwner(ctx, req.OwnerID, req.OwnerType)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find queue for owner")
	}

	// Enqueue the signal
	enqueueResp, err := a.queueClient.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  req.Signal,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to enqueue signal")
	}

	return &EnqueueSignalToOwnerResponse{
		QueueSignalID: enqueueResp.ID,
		WorkflowID:    enqueueResp.WorkflowID,
	}, nil
}
