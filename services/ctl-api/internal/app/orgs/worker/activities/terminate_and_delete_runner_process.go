package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type TerminateAndDeleteRunnerProcessRequest struct {
	ProcessID string `validate:"required"`
	RunnerID  string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) TerminateAndDeleteRunnerProcess(ctx context.Context, req TerminateAndDeleteRunnerProcessRequest) error {
	if err := a.runnersHelpers.TerminateProcessQueueStrict(ctx, req.RunnerID, req.ProcessID); err != nil {
		return errors.Wrap(err, "unable to terminate process queue")
	}

	if res := a.db.WithContext(ctx).
		Delete(&app.RunnerProcess{}, "id = ?", req.ProcessID); res.Error != nil {
		return errors.Wrap(res.Error, "unable to delete runner process")
	}

	return nil
}
