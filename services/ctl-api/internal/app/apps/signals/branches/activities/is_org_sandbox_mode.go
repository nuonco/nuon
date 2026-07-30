package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type IsOrgSandboxModeRequest struct {
	OrgID string `json:"org_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) IsOrgSandboxMode(ctx context.Context, req IsOrgSandboxModeRequest) (bool, error) {
	var org app.Org
	if err := a.db.WithContext(ctx).First(&org, "id = ?", req.OrgID).Error; err != nil {
		return false, fmt.Errorf("unable to get org: %w", err)
	}
	return org.SandboxMode, nil
}
