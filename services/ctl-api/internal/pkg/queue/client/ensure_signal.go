package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type EnsureSignalRequest struct {
	OwnerID    string            `json:"owner_id" validate:"required"`
	OwnerType  string            `json:"owner_type" validate:"required"`
	SignalType signal.SignalType `json:"signal_type" validate:"required"`

	// Callback to register on the signal if it is still in flight.
	Callback callback.Ref `json:"callback"`
}

type EnsureSignalResponse struct {
	// AlreadyComplete is true if the signal has already finished successfully.
	AlreadyComplete bool   `json:"already_complete"`
	QueueSignalID   string `json:"queue_signal_id"`
}

// EnsureSignal checks the latest signal of the given type for the owner.
// If the signal is already complete (success), it returns immediately.
// If the signal is in flight (queued/in_progress), it atomically appends
// the callback and returns AlreadyComplete=false so the caller can Await.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) EnsureSignal(ctx context.Context, req *EnsureSignalRequest) (*EnsureSignalResponse, error) {
	// Find the latest signal for this owner+type.
	var qs app.QueueSignal
	res := c.db.WithContext(ctx).
		Where(app.QueueSignal{
			OwnerID:   req.OwnerID,
			OwnerType: req.OwnerType,
			Type:      req.SignalType,
		}).
		Order("created_at DESC").
		First(&qs)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "no signal found for owner+type")
	}

	// Already succeeded — nothing to wait for.
	if qs.Status.Status == app.StatusSuccess {
		return &EnsureSignalResponse{AlreadyComplete: true, QueueSignalID: qs.ID}, nil
	}

	// Still in flight — atomically append the callback.
	if qs.Status.Status == app.StatusQueued || qs.Status.Status == app.StatusInProgress {
		if req.Callback.IsSet() {
			cbJSON, err := json.Marshal([]callback.Ref{req.Callback})
			if err != nil {
				return nil, errors.Wrap(err, "unable to marshal callback")
			}
			if err := c.db.WithContext(ctx).Exec(
				`UPDATE queue_signals SET callbacks = COALESCE(callbacks, '[]'::jsonb) || ?::jsonb WHERE id = ?`,
				string(cbJSON), qs.ID,
			).Error; err != nil {
				return nil, errors.Wrap(err, "unable to add callback to signal")
			}
		}

		// Re-check status after append to handle the race where the signal
		// completed between our initial read and the callback append.
		var recheck app.QueueSignal
		if err := c.db.WithContext(ctx).First(&recheck, "id = ?", qs.ID).Error; err == nil {
			if recheck.Status.Status == app.StatusSuccess {
				return &EnsureSignalResponse{AlreadyComplete: true, QueueSignalID: qs.ID}, nil
			}
		}

		return &EnsureSignalResponse{AlreadyComplete: false, QueueSignalID: qs.ID}, nil
	}

	// Terminal non-success (error, cancelled, etc.)
	return nil, fmt.Errorf("signal %s is in terminal non-success state: %s - %s",
		qs.ID, qs.Status.Status, qs.Status.StatusHumanDescription)
}
