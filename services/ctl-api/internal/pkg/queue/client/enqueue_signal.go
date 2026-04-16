package client

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/pkg/errors"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

type EnqueueSignalRequest struct {
	QueueID   string        `validate:"required"`
	Signal    signal.Signal `validate:"required"`
	OwnerID   string
	OwnerType string
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) EnqueueSignal(ctx context.Context, req *EnqueueSignalRequest) (*queue.EnqueueResponse, error) {
	q, err := c.getQueue(ctx, req.QueueID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue")
	}

	// Create the QueueSignal record in the DB directly so we can return the
	// signal ID without waiting for the queue workflow to process it.
	suffix := make([]byte, 3)
	_, _ = rand.Read(suffix)

	queueSignal := app.QueueSignal{
		Signal: signaldb.SignalData{
			Signal: req.Signal,
		},
		QueueID:   req.QueueID,
		Type:      req.Signal.Type(),
		OwnerID:   req.OwnerID,
		OwnerType: req.OwnerType,
		Status:    app.NewCompositeStatus(ctx, app.StatusQueued),
		Workflow: signaldb.WorkflowRef{
			Namespace:  q.Workflow.Namespace,
			IDTemplate: q.Workflow.ID + "-handler-%s-" + string(req.Signal.Type()) + "-" + hex.EncodeToString(suffix),
		},
	}

	if res := c.db.WithContext(ctx).Create(&queueSignal); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create queue signal")
	}

	// Send the enqueue update to the queue workflow. We only wait for the
	// "accepted" stage so the caller gets the signal ID back immediately.
	_, err = c.tClient.UpdateWithStartWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWithStartWorkflowOptions{
		UpdateOptions: tclient.UpdateWorkflowOptions{
			WorkflowID:   q.Workflow.ID,
			UpdateName:   queue.EnqueueUpdateName,
			WaitForStage: tclient.WorkflowUpdateStageAccepted,
			Args: []any{
				queue.EnqueueHandlerInput{
					QueueSignalID: queueSignal.ID,
					WorkflowID:    queueSignal.Workflow.ID,
				},
			},
		},
		StartWorkflowOperation: c.queueStartOperation(q),
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to call enqueue handler")
	}

	return &queue.EnqueueResponse{
		ID:         queueSignal.ID,
		WorkflowID: queueSignal.Workflow.ID,
	}, nil
}
