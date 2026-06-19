package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type ClearInstallStackVersionCallbackRequest struct {
	VersionID string `json:"version_id" validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) ClearInstallStackVersionCallback(ctx context.Context, req ClearInstallStackVersionCallbackRequest) error {
	if res := a.db.WithContext(ctx).
		Model(&app.InstallStackVersion{}).
		Where("id = ?", req.VersionID).
		Update("callback_ref", callback.Ref{}); res.Error != nil {
		return generics.TemporalGormError(res.Error, "unable to clear callback ref")
	}
	return nil
}
