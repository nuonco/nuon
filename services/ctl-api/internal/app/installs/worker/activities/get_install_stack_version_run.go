package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type GetInstallStackVersionRunRequest struct {
	VersionID string `json:"version_id" validate:"required"`
}

// @temporal-gen activity
// @by-id VersionID
func (a *Activities) GetInstallStackVersionRun(ctx context.Context, req GetInstallStackVersionRunRequest) (*app.InstallStackVersionRun, error) {
	var stack app.InstallStackVersionRun

	if res := a.db.WithContext(ctx).
		Where(app.InstallStackVersionRun{
			InstallStackVersionID: req.VersionID,
		}).
		Order("created_at DESC").
		First(&stack); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to get install stack version run")
	}

	return &stack, nil
}
