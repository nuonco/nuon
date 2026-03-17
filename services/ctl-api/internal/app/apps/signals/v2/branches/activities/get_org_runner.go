package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetOrgRunnerRequest struct {
	OrgID string `json:"org_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) GetOrgRunner(ctx context.Context, req GetOrgRunnerRequest) (*app.Runner, error) {
	var rg app.RunnerGroup
	res := a.db.WithContext(ctx).
		Preload("Runners").
		Where("owner_type = ? AND owner_id = ?", "orgs", req.OrgID).
		First(&rg)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner group for org %s: %w", req.OrgID, res.Error)
	}
	if len(rg.Runners) == 0 {
		return nil, fmt.Errorf("no runners found for org %s", req.OrgID)
	}
	return &rg.Runners[0], nil
}
