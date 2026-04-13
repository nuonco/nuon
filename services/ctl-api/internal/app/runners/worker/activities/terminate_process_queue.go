package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type TerminateProcessQueueRequest struct {
	RunnerID  string `validate:"required"`
	ProcessID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field RunnerID
func (a *Activities) TerminateProcessQueue(ctx context.Context, req TerminateProcessQueueRequest) error {
	if err := a.helpers.TerminateProcessQueue(ctx, req.RunnerID, req.ProcessID); err != nil {
		return generics.TemporalGormError(err, "unable to terminate process queue")
	}
	return nil
}
