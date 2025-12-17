package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type GetInstallForSandboxRequest struct {
	SandboxID string `json:"sandbox_id" validate:"required"`
}

// @temporal-gen activity
// @by-id SandboxID
func (a *Activities) GetInstallForSandbox(ctx context.Context, req GetInstallForSandboxRequest) (*app.Install, error) {
	var sandbox app.InstallSandbox

	res := a.db.WithContext(ctx).
		Where(app.InstallSandbox{
			ID: req.SandboxID,
		}).
		First(&sandbox)
	if res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to get install sandbox")
	}

	return a.getInstall(ctx, sandbox.InstallID)
}
