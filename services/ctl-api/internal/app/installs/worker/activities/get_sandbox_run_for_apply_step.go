package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetSandboxRunForApplyRequest struct {
	InstallID         string `validate:"required"`
	InstallWorkflowID string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) GetSandboxRunForApply(ctx context.Context, req GetSandboxRunForApplyRequest) (*app.InstallSandboxRun, error) {
	var run app.InstallSandboxRun

	res := a.db.WithContext(ctx).
		Where(app.InstallSandboxRun{
			InstallID:         req.InstallID,
			InstallWorkflowID: generics.ToPtr(req.InstallWorkflowID),
		}).
		First(&run)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install sandbox run")
	}

	return &run, nil
}
