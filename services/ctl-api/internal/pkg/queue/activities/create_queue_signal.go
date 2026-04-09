package activities

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"go.temporal.io/sdk/activity"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

type CreateQueueSignalRequest struct {
	QueueID string        `json:"queue_id" validate:"required"`
	Signal  signal.Signal `json:"signal" validate:"required"`

	// OwnerID and OwnerType are optional — when set they populate the polymorphic
	// owner association on the created QueueSignal so no separate UPDATE is needed.
	OwnerID   string `json:"owner_id,omitempty"`
	OwnerType string `json:"owner_type,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateQueueSignal(ctx context.Context, req *CreateQueueSignalRequest) (*app.QueueSignal, error) {
	info := activity.GetInfo(ctx)

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
		Workflow: signaldb.WorkflowRef{
			Namespace:  info.WorkflowNamespace,
			IDTemplate: info.WorkflowExecution.ID + "-handler-%s-" + string(req.Signal.Type()) + "-" + hex.EncodeToString(suffix),
		},
	}

	if res := a.db.WithContext(ctx).Create(&queueSignal); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to create queue signal")
	}

	return &queueSignal, nil
}
