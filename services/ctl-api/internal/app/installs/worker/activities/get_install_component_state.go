package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

type GetInstallComponentStateRequest struct {
	InstallComponentID string `validate:"required"`
}

// @temporal-gen activity
// @by-id InstallComponentID
func (h *Activities) GetInstallComponentState(ctx context.Context, req GetInstallComponentStateRequest) (*app.InstallComponent, error) {
	var installComponent app.InstallComponent

	res := h.db.WithContext(ctx).
		Preload("InstallDeploys", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(
				scopes.WithOverrideTable(views.CustomViewName(db, &app.InstallDeploy{}, "state_view_v1")),
			)
		}).
		Preload("InstallDeploys.RunnerJobs", func(db *gorm.DB) *gorm.DB {
			return db.Order(views.TableOrViewName(db, &app.RunnerJob{}, ".created_at DESC"))
		}).
		Preload("Component").
		Preload("TerraformWorkspace").
		First(&installComponent, "id = ?", req.InstallComponentID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install components: %w", res.Error)
	}

	return &installComponent, nil
}
