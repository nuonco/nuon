package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/signal/db"
)

// NOTE(jm): temporal-gen does not support activities with multiple args
func AwaitCreateQueueSignal(ctx workflow.Context, sig signal.Signal, req *CreateQueueSignalRequest) (*app.QueueSignal, error) {
	_ = (&Activities{}).CreateQueueSignal
	// use this ^ for to go-to-definition jumping in your editor

	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 30 * time.Minute,
		StartToCloseTimeout:    5 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// The func pointer being passed here is not used for execution. Temporal uses it as the
	// key for a name lookup against its activity registry, so that it knows what activity function
	// to actually call when the workflow is ready to be executed.
	fut := workflow.ExecuteActivity(ctx, (&Activities{}).CreateQueueSignal, sig, req)
	var ret app.QueueSignal
	if err := fut.Get(ctx, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}

type CreateQueueSignalRequest struct {
	QueueID string `validate:"required"`
}

func (a *Activities) CreateQueueSignal(ctx context.Context, sig signal.Signal, req *CreateQueueSignalRequest) (*app.QueueSignal, error) {
	return a.createQueueSignal(ctx, req.QueueID, sig)
}

func (a *Activities) createQueueSignal(ctx context.Context, queueID string, signal signal.Signal) (*app.QueueSignal, error) {
	info := activity.GetInfo(ctx)

	queueSignal := app.QueueSignal{
		Signal: signaldb.SignalData{
			Signal: signal,
		},
		QueueID: queueID,
		Type:    signal.Type(),
		Workflow: signaldb.WorkflowRef{
			Namespace:  info.WorkflowNamespace,
			IDTemplate: info.WorkflowExecution.ID + "-handler-%s",
		},
	}

	if res := a.db.WithContext(ctx).
		Create(&queueSignal); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create queue")
	}

	return &queueSignal, nil
}
