package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetRunnerProcessRequest struct {
	ProcessID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field ProcessID
func (a *Activities) GetRunnerProcess(ctx context.Context, req GetRunnerProcessRequest) (*app.RunnerProcess, error) {
	var process app.RunnerProcess
	res := a.db.WithContext(ctx).
		Preload("Shutdowns").
		First(&process, "id = ?", req.ProcessID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner process: %w", res.Error)
	}

	return &process, nil
}
