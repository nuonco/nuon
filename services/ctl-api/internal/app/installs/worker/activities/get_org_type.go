package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetOrgTypeRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallID
func (a *Activities) GetOrgType(ctx context.Context, req GetOrgRequest) (app.OrgType, error) {
	return a.features.OrgType(ctx)
}
