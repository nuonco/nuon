package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
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
		return nil, dbgenerics.TemporalGormError(res.Error, "unable to get runner process")
	}

	return &process, nil
}
