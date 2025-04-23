package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetSandboxRunRequest struct {
	RunID string `validate:"required"`
}

// @temporal-gen activity
// @by-id RunID
func (a *Activities) GetSandboxRun(ctx context.Context, req GetSandboxRunRequest) (*app.InstallSandboxRun, error) {
	var run app.InstallSandboxRun

	res := a.db.WithContext(ctx).
		First(&run, "id = ?", req.RunID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install sandbox run")
	}

	return &run, nil
}
