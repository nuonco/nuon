package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type GetInstallStackVersionByIDRequest struct {
	VersionID string `json:"version_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field VersionID
func (a *Activities) GetInstallStackVersionByID(ctx context.Context, req GetInstallStackVersionByIDRequest) (*app.InstallStackVersion, error) {
	var version app.InstallStackVersion
	if res := a.db.WithContext(ctx).Where("id = ?", req.VersionID).First(&version); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to get install stack version")
	}
	return &version, nil
}
