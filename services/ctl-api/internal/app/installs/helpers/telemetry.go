package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

func (h *Helpers) SetInstallTelemetry(ctx context.Context, installID string, enabled *bool) error {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return err
	}
	var install struct{ ID string }
	if err := h.db.WithContext(ctx).Model(&app.Install{}).Scopes(scopes.WithDisableViews).
		Where(app.Install{ID: installID, OrgID: orgID}).Take(&install).Error; err != nil {
		return fmt.Errorf("get install for telemetry update: %w", err)
	}
	if enabled == nil {
		return h.db.WithContext(ctx).Model(&app.InstallConfig{}).
			Where(app.InstallConfig{InstallID: installID, OrgID: orgID}).
			Update("telemetry_enabled", nil).Error
	}
	return h.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "install_id"}, {Name: "deleted_at"}},
		DoUpdates: clause.AssignmentColumns([]string{"telemetry_enabled", "updated_at"}),
	}).Create(&app.InstallConfig{InstallID: installID, TelemetryEnabled: enabled}).Error
}
