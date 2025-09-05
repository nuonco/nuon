package activities

import (
	"context"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

type GetInstallStackVersionRequest struct {
	InstallID string `json:"id" validate:"required"`
}

// @temporal-gen activity
// @by-id InstallID
func (a *Activities) GetInstallStackVersion(ctx context.Context, req GetInstallStackRequest) (*app.InstallStackVersion, error) {
	var stack app.InstallStack

	if res := a.db.WithContext(ctx).
		Where(app.InstallStack{
			InstallID: req.InstallID,
		}).
		Preload("InstallStackVersions", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_stack_versions.created_at DESC").Limit(1)
		}).
		First(&stack); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error)
	}

	if len(stack.InstallStackVersions) < 1 {
		return nil, generics.TemporalGormError(gorm.ErrRecordNotFound)
	}

	return &stack.InstallStackVersions[0], nil
}
