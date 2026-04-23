package callback

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	tclient "go.temporal.io/sdk/client"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ActivitiesParams struct {
	fx.In

	DB      *gorm.DB `name:"psql"`
	TClient temporalclient.Client
}

type Activities struct {
	db      *gorm.DB
	tClient temporalclient.Client
}

func NewActivities(params ActivitiesParams) *Activities {
	return &Activities{
		db:      params.DB,
		tClient: params.TClient,
	}
}

type InvokeCallbacksRequest struct {
	QueueSignalID string `json:"queue_signal_id" validate:"required"`
	QueueID       string `json:"queue_id" validate:"required"`
	Event         Event  `json:"event" validate:"required"`
	ErrMessage    string `json:"err_message,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) InvokeCallbacks(ctx context.Context, req *InvokeCallbacksRequest) error {
	if req == nil {
		return fmt.Errorf("invoke callbacks request is nil")
	}

	l := temporalzap.GetActivityLogger(ctx).With(
		zap.String("queue_signal_id", req.QueueSignalID),
		zap.String("queue_id", req.QueueID),
		zap.String("event", string(req.Event)),
	)

	var callbacks []app.QueueSignalCallback
	res := a.db.WithContext(ctx).Where(app.QueueSignalCallback{
		QueueSignalID: req.QueueSignalID,
		Event:         string(req.Event),
	}).Find(&callbacks)
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to query callbacks")
	}

	if len(callbacks) == 0 {
		return nil
	}

	l.Info("invoking callbacks", zap.Int("count", len(callbacks)))

	payload := CallbackPayload{
		Event:         req.Event,
		QueueSignalID: req.QueueSignalID,
		QueueID:       req.QueueID,
		ErrMessage:    req.ErrMessage,
	}

	for _, cb := range callbacks {
		uh := cb.UpdateHandler
		_, err := a.tClient.UpdateWorkflowInNamespace(ctx, uh.Namespace, tclient.UpdateWorkflowOptions{
			WorkflowID:   uh.WorkflowID,
			UpdateName:   uh.UpdateName,
			WaitForStage: tclient.WorkflowUpdateStageAccepted,
			Args:         []any{payload},
		})
		if err != nil {
			l.Warn("callback invocation failed",
				zap.String("callback_id", cb.ID),
				zap.String("workflow_id", uh.WorkflowID),
				zap.String("update_name", uh.UpdateName),
				zap.Error(err))
			continue
		}
	}

	return nil
}
