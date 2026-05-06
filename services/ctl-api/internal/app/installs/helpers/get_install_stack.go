package helpers

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// GetInstallStack returns the install stack row for the given install.
// Exported so service handlers can resolve InstallStackID without
// re-implementing the lookup.
func (h *Helpers) GetInstallStack(ctx context.Context, installID string) (*app.InstallStack, error) {
	return h.getInstallStack(ctx, installID)
}

// getInstallStacks gets an install stack.
func (h *Helpers) getInstallStack(ctx context.Context, installID string) (*app.InstallStack, error) {
	var installStack app.InstallStack
	res := h.db.WithContext(ctx).
		Preload("InstallStackOutputs").
		Preload("InstallStackVersions", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_stack_versions.created_at DESC")
		}).
		Where(app.InstallStack{
			InstallID: installID,
		}).
		Find(&installStack)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install state")
	}

	return &installStack, nil
}
