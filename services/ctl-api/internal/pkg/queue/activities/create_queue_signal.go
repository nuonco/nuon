package activities

import (
	"context"

	"github.com/pkg/errors"

	"go.temporal.io/sdk/activity"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// @temporal-gen-v2 activity
// @as-wrapper
// @by-field QueueID
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
