package client

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

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

	status := app.NewCompositeStatus(ctx, app.StatusQueued)
	if t, ok := req.Signal.(signal.SignalWithTimeout); ok {
		if status.Metadata == nil {
			status.Metadata = make(map[string]any)
		}
		status.Metadata["timeout_ns"] = t.Timeout().Nanoseconds()
	}

	queueSignal := app.QueueSignal{
		Signal: signaldb.SignalData{
			Signal: req.Signal,
		},
		QueueID:   req.QueueID,
		Type:      req.Signal.Type(),
		OwnerID:   req.OwnerID,
		OwnerType: req.OwnerType,
		Status:    status,
		Workflow: signaldb.WorkflowRef{
			Namespace:  q.Workflow.Namespace,
			IDTemplate: q.Workflow.ID + "-handler-%s-" + string(req.Signal.Type()) + "-" + hex.EncodeToString(suffix),
		},
	}

	if res := c.db.WithContext(ctx).Create(&queueSignal); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create queue signal")
	}

	// Fire off the UpdateWithStart call in a background goroutine so the caller
	// gets the signal ID back immediately without waiting for the queue workflow.
	enqueueStartedAt := time.Now().UTC().Format(time.RFC3339)
	startOp := c.queueStartOperation(q)
	namespace := q.Workflow.Namespace
	workflowID := q.Workflow.ID
	signalID := queueSignal.ID
	handlerWorkflowID := queueSignal.Workflow.ID

	go func() {
		bgCtx := context.Background()
		_, err := c.tClient.UpdateWithStartWorkflowInNamespace(bgCtx, namespace, tclient.UpdateWithStartWorkflowOptions{
			UpdateOptions: tclient.UpdateWorkflowOptions{
				WorkflowID:   workflowID,
				UpdateName:   queue.EnqueueUpdateName,
				WaitForStage: tclient.WorkflowUpdateStageAccepted,
				Args: []any{
					queue.EnqueueHandlerInput{
						QueueSignalID: signalID,
						WorkflowID:    handlerWorkflowID,
					},
				},
			},
			StartWorkflowOperation: startOp,
		})

		metadata := map[string]any{
			"enqueue_started_at":  enqueueStartedAt,
			"enqueue_finished_at": time.Now().UTC().Format(time.RFC3339),
		}
		if err != nil {
			metadata["enqueue_error"] = err.Error()
			c.l.Warn("background enqueue failed",
				zap.String("queue-signal-id", signalID),
				zap.Error(err))
		}
		c.updateQueueSignalMetadata(signalID, metadata)
	}()

	return &queue.EnqueueResponse{
		ID:         queueSignal.ID,
		WorkflowID: queueSignal.Workflow.ID,
	}, nil
}
