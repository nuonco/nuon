package activities

import (
	"context"
	"errors"
	"fmt"

	"go.temporal.io/sdk/temporal"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetCurrentRunnerProcessRequest struct {
	RunnerID    string                `validate:"required"`
	ProcessType app.RunnerProcessType `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field RunnerID
func (a *Activities) GetCurrentRunnerProcess(ctx context.Context, req GetCurrentRunnerProcessRequest) (*app.RunnerProcess, error) {
	var process app.RunnerProcess
	res := a.db.WithContext(ctx).
		Where("runner_id = ? AND type = ?", req.RunnerID, req.ProcessType).
		Order("created_at DESC").
		First(&process)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, temporal.NewNonRetryableApplicationError("not found", "not found", res.Error, "")
		}

		return nil, fmt.Errorf("unable to get current runner process: %w", res.Error)
	}

	return &process, nil
}
