package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type EnqueueSignalToOwnerRequest struct {
	OwnerID   string        `json:"owner_id" validate:"required"`
	OwnerType string        `json:"owner_type" validate:"required"`
	QueueName string        `json:"queue_name,omitempty"`
	Signal    signal.Signal `json:"signal" validate:"required"`

	// QueueID short-circuits the owner/name lookup when set.
	QueueID string `json:"queue_id,omitempty"`

	// SignalOwnerID and SignalOwnerType are set on the QueueSignal record to track
	// which entity (e.g. workflow step) triggered this signal execution.
	SignalOwnerID   string `json:"signal_owner_id,omitempty"`
	SignalOwnerType string `json:"signal_owner_type,omitempty"`
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

	// Resolve queue ID — use direct ID if provided, otherwise look up by owner.
	queueID := req.QueueID
	if queueID == "" {
		var queue *app.Queue
		var err error
		if req.QueueName != "" {
			queue, err = a.queueClient.GetQueueByOwnerAndName(ctx, req.OwnerID, req.OwnerType, req.QueueName)
		} else {
			queue, err = a.queueClient.GetQueueByOwner(ctx, req.OwnerID, req.OwnerType)
		}
		if err != nil {
			return nil, errors.Wrap(err, "unable to find queue for owner")
		}
		queueID = queue.ID
	}

	// Enqueue the signal
	enqueueResp, err := a.queueClient.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID:   queueID,
		Signal:    req.Signal,
		OwnerID:   req.SignalOwnerID,
		OwnerType: req.SignalOwnerType,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to enqueue signal")
	}

	return &EnqueueSignalToOwnerResponse{
		QueueSignalID: enqueueResp.ID,
		WorkflowID:    enqueueResp.WorkflowID,
	}, nil
}
