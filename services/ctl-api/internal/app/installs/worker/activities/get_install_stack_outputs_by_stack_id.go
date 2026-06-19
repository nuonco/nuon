package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type GetInstallStackOutputsByStackIDRequest struct {
	InstallStackID string `json:"install_stack_id" validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) GetInstallStackOutputsByStackID(ctx context.Context, req GetInstallStackOutputsByStackIDRequest) (*app.InstallStackOutputs, error) {
	var outputs app.InstallStackOutputs
	if res := a.db.WithContext(ctx).
		Where(app.InstallStackOutputs{InstallStackID: req.InstallStackID}).
		First(&outputs); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to get install stack outputs")
	}

	return &outputs, nil
}
