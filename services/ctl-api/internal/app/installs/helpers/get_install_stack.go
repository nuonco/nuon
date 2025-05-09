package helpers

import (
	"context"

	"gorm.io/gorm"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// getInstallStacks gets an install stack.
func (h *Helpers) getInstallStack(ctx context.Context, installID string) (*app.InstallStack, error) {
	var installStack app.InstallStack
	res := h.db.WithContext(ctx).
		Preload("InstallStackVersions", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(
				scopes.WithOverrideTable(views.CustomViewName(db, &app.InstallStackVersionRun{}, "state_view_v1")),
			)
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
