package activities

import (
	"context"

	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
)

type MigrateOrgHealthcheckSweepsRequest struct {
	OrgID   string `validate:"required"`
	Enabled bool   `json:"enabled"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (a *Activities) MigrateOrgHealthcheckSweeps(ctx context.Context, req MigrateOrgHealthcheckSweepsRequest) (*runnershelpers.OrgHealthcheckMigrationResult, error) {
	if req.Enabled {
		return a.runnersHelpers.MigrateOrgToHealthcheckSweeps(ctx, req.OrgID)
	}
	return a.runnersHelpers.MigrateOrgFromHealthcheckSweeps(ctx, req.OrgID)
}
