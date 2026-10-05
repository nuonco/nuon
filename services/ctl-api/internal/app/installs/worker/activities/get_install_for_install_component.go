package activities

import (
	"context"

	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type GetInstallForInstallComponentRequest struct {
	InstallComponentID string `json:"component_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallComponentID
func (a *Activities) GetInstallForInstallComponent(ctx context.Context, req GetInstallForInstallComponentRequest) (*app.Install, error) {
	if req.InstallComponentID == "" {
		return nil, temporal.NewNonRetryableApplicationError("install component id is required", "InvalidArgument", nil)
	}

	var component app.InstallComponent

	res := a.db.WithContext(ctx).
		First(&component, "id = ?", req.InstallComponentID)
	if res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to get install component")
	}

	return a.getInstall(ctx, component.InstallID)
}
