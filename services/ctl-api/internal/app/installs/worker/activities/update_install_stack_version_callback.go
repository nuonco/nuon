package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type UpdateInstallStackVersionCallbackRequest struct {
	VersionID   string       `json:"version_id" validate:"required"`
	CallbackRef callback.Ref `json:"callback_ref"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateInstallStackVersionCallback(ctx context.Context, req UpdateInstallStackVersionCallbackRequest) error {
	res := a.db.WithContext(ctx).
		Model(&app.InstallStackVersion{}).
		Where("id = ?", req.VersionID).
		Update("callback_ref", req.CallbackRef)
	if res.Error != nil {
		return generics.TemporalGormError(res.Error, "unable to update callback ref")
	}
	return nil
}
