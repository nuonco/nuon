package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type UpdateQueueSignalStatusRequest struct {
	QueueSignalID     string         `json:"queue_signal_id" validate:"required"`
	Status            app.Status     `json:"status" validate:"required"`
	StatusDescription string         `json:"status_description,omitempty"`
	Metadata          map[string]any `json:"metadata,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) UpdateQueueSignalStatus(ctx context.Context, req *UpdateQueueSignalStatusRequest) error {
	// Read existing status to preserve history and metadata
	var existing app.QueueSignal
	if err := a.db.WithContext(ctx).
		Select("status").
		Where("id = ?", req.QueueSignalID).
		First(&existing).Error; err != nil {
		return generics.TemporalGormError(err, fmt.Sprintf("unable to read queue signal %s", req.QueueSignalID))
	}

	cs := app.NewCompositeStatus(ctx, req.Status)
	if req.StatusDescription != "" {
		cs.StatusHumanDescription = req.StatusDescription
	}
	for k, v := range req.Metadata {
		cs.Metadata[k] = v
	}

	// Carry forward metadata from existing status that isn't being overwritten
	for k, v := range existing.Status.Metadata {
		if _, ok := cs.Metadata[k]; !ok {
			cs.Metadata[k] = v
		}
	}

	// Build history: existing history + existing status (without nested history)
	history := existing.Status.History
	existing.Status.History = nil
	cs.History = append(history, existing.Status)

	res := a.db.WithContext(ctx).
		Model(&app.QueueSignal{}).
		Where("id = ?", req.QueueSignalID).
		Update("status", cs)
	if res.Error != nil {
		return generics.TemporalGormError(res.Error, "unable to update queue signal status")
	}

	return nil
}
