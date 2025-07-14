package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
	"gorm.io/gorm"
)

type GetInstallComponentIDsRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen activity
// @by-id InstallID
func (a *Activities) GetInstallComponentIDs(ctx context.Context, req GetInstallComponentIDsRequest) ([]string, error) {
	install := &app.Install{}
	res := a.db.WithContext(ctx).
		Preload("InstallComponents", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_components.created_at DESC")
		}).
		First(&install, "id = ?", req.InstallID)
	if res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "get install components")
	}

	ids := make([]string, 0, len(install.InstallComponents))
	for _, comp := range install.InstallComponents {
		ids = append(ids, comp.ID)
	}
	return ids, nil
}
