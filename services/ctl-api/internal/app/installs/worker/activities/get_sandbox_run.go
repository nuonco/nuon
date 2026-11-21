package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetSandboxRunRequest struct {
	RunID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field RunID
func (a *Activities) GetSandboxRun(ctx context.Context, req GetSandboxRunRequest) (*app.InstallSandboxRun, error) {
	var run app.InstallSandboxRun

	res := a.db.WithContext(ctx).
		Preload("Install").
		Preload("InstallWorkflow").
		Preload("AppSandboxConfig").
		First(&run, "id = ?", req.RunID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install sandbox run")
	}
	if run.InstallWorkflow != nil {
		if _, err := run.InstallWorkflow.PreparedInstallConfigID(run.Install); err != nil {
			return nil, err
		}
	}

	return &run, nil
}
