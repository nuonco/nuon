package activities

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetInstallStackStateRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen activity
// @by-id InstallID
func (a *Activities) GetInstallStackState(ctx context.Context, req GetInstallStackRequest) (*app.InstallStack, error) {
	var installStack app.InstallStack
	res := a.db.WithContext(ctx).
		Preload("InstallStackOutputs").
		Preload("InstallStackVersions", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_stack_versions.created_at DESC")
		}).
		Where(app.InstallStack{
			InstallID: req.InstallID,
		}).
		Find(&installStack)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install state")
	}

	return &installStack, nil
}
