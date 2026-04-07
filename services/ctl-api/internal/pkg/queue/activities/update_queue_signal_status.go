package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateQueueSignalStatusRequest struct {
	QueueSignalID     string     `json:"queue_signal_id" validate:"required"`
	Status            app.Status `json:"status" validate:"required"`
	StatusDescription string     `json:"status_description,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) UpdateQueueSignalStatus(ctx context.Context, req *UpdateQueueSignalStatusRequest) error {
	cs := app.NewCompositeStatus(ctx, req.Status)
	if req.StatusDescription != "" {
		cs.StatusHumanDescription = req.StatusDescription
	}

	res := a.db.WithContext(ctx).
		Model(&app.QueueSignal{}).
		Where("id = ?", req.QueueSignalID).
		Update("status", cs)
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update queue signal status")
	}

	return nil
}
