package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

func (h *Helpers) GetInstallComponent(ctx context.Context, installID, componentID string) (*app.InstallComponent, error) {
	installCmp := app.InstallComponent{}
	res := h.db.WithContext(ctx).
		Preload("InstallDeploys", func(db *gorm.DB) *gorm.DB {
			return db.
				Scopes(scopes.WithOverrideTable(views.CustomViewName(db, &app.InstallDeploy{}, "state_view_v1")))
		}).
		Preload("TerraformWorkspace").
		Where(&app.InstallComponent{
			InstallID:   installID,
			ComponentID: componentID,
		}).
		First(&installCmp)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install component: %w", res.Error)
	}

	return &installCmp, nil
}
