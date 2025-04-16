package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetOrgTypeRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen activity
// @by-id InstallID
func (a *Activities) GetOrgType(ctx context.Context, req GetOrgRequest) (app.OrgType, error) {
	return a.features.OrgType(ctx)
}
