package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type UpdateQueueSignalRunIDRequest struct {
	QueueSignalID string `json:"queue_signal_id" validate:"required"`
	RunID         string `json:"run_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 10s
func (a *Activities) UpdateQueueSignalRunID(ctx context.Context, req *UpdateQueueSignalRunIDRequest) error {
	var qs app.QueueSignal
	if res := a.db.WithContext(ctx).First(&qs, "id = ?", req.QueueSignalID); res.Error != nil {
		return generics.TemporalGormError(res.Error, fmt.Sprintf("unable to get queue signal %s", req.QueueSignalID))
	}

	// Write-once: only set the RunID if it hasn't been set yet.
	// This preserves the original handler run's ID so that callers like
	// FetchSteps always target the run that actually has the in-memory results.
	if qs.Workflow.RunID != "" {
		return nil
	}

	qs.Workflow.RunID = req.RunID
	res := a.db.WithContext(ctx).
		Model(&app.QueueSignal{}).
		Where("id = ?", req.QueueSignalID).
		Update("workflow", qs.Workflow)
	if res.Error != nil {
		return generics.TemporalGormError(res.Error, fmt.Sprintf("unable to update run_id for queue signal %s", req.QueueSignalID))
	}

	return nil
}
