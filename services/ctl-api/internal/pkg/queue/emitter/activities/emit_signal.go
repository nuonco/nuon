package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
)

type EmitSignalRequest struct {
	EmitterID string `validate:"required"`
	QueueID   string `validate:"required"`
}

type EmitSignalResponse struct {
	QueueSignalID string
	WorkflowID    string
}

func AwaitEmitSignal(ctx workflow.Context, req *EmitSignalRequest) (*EmitSignalResponse, error) {
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 2 * time.Minute,
		StartToCloseTimeout:    30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	fut := workflow.ExecuteActivity(ctx, (&Activities{}).EmitSignal, req)
	var ret EmitSignalResponse
	if err := fut.Get(ctx, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}

// @temporal-gen activity
func (a *Activities) EmitSignal(ctx context.Context, req *EmitSignalRequest) (*EmitSignalResponse, error) {
	// Get the emitter to access its signal template
	var emitter app.QueueEmitter
	if res := a.db.WithContext(ctx).
		Where("id = ?", req.EmitterID).
		First(&emitter); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get emitter")
	}

	if emitter.SignalTemplate.Signal == nil {
		return nil, errors.New("emitter has no signal template configured")
	}

	// Get the queue to find its workflow details
	var q app.Queue
	if res := a.db.WithContext(ctx).
		Where("id = ?", req.QueueID).
		First(&q); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get queue")
	}

	// Call the queue's enqueue update handler
	rawResp, err := a.tClient.UpdateWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWorkflowOptions{
		WorkflowID:   q.Workflow.ID,
		UpdateName:   queue.EnqueueUpdateName,
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
		Args: []any{
			emitter.SignalTemplate.Signal,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to enqueue signal to queue")
	}

	var enqueueResp queue.EnqueueResponse
	if err := rawResp.Get(ctx, &enqueueResp); err != nil {
		return nil, errors.Wrap(err, "unable to get enqueue response")
	}

	a.l.Info("signal emitted to queue",
		zap.String("emitter-id", req.EmitterID),
		zap.String("queue-id", req.QueueID),
		zap.String("queue-signal-id", enqueueResp.ID),
		zap.String("workflow-id", enqueueResp.WorkflowID),
	)

	return &EmitSignalResponse{
		QueueSignalID: enqueueResp.ID,
		WorkflowID:    enqueueResp.WorkflowID,
	}, nil
}
