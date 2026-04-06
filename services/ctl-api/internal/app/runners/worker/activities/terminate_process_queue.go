package activities

import (
	"context"
)

type TerminateProcessQueueRequest struct {
	RunnerID  string `validate:"required"`
	ProcessID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field RunnerID
func (a *Activities) TerminateProcessQueue(ctx context.Context, req TerminateProcessQueueRequest) error {
	return a.helpers.TerminateProcessQueue(ctx, req.RunnerID, req.ProcessID)
}
