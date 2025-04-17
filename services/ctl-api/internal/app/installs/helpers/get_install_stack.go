package helpers

import (
	"context"

	"gorm.io/gorm"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// getInstallStacks gets an install stack.
func (h *Helpers) getInstallStack(ctx context.Context, installID string) (*app.InstallStack, error) {
	var installStack app.InstallStack
	res := h.db.WithContext(ctx).
		Preload("InstallStackVersions", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_stack_versions.created_at DESC").Limit(1)
		}).
		Preload("InstallStackOutputs").
		Where(app.InstallStack{
			InstallID: installID,
		}).
		Find(&installStack)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install state")
	}

	return &installStack, nil
}
