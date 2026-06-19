package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type GetInstallStackVersionCallbackRequest struct {
	VersionID string `json:"version_id" validate:"required"`
}

type GetInstallStackVersionCallbackResponse struct {
	CallbackRef callback.Ref `json:"callback_ref"`
}

// @temporal-gen-v2 activity
func (a *Activities) GetInstallStackVersionCallback(ctx context.Context, req GetInstallStackVersionCallbackRequest) (*GetInstallStackVersionCallbackResponse, error) {
	var version app.InstallStackVersion
	if res := a.db.WithContext(ctx).
		Select("id", "callback_ref").
		Where("id = ?", req.VersionID).
		First(&version); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to get version callback")
	}
	return &GetInstallStackVersionCallbackResponse{CallbackRef: version.CallbackRef}, nil
}
