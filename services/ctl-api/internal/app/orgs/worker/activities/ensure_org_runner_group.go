package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type EnsureOrgRunnerGroupRequest struct {
	OrgID string `validate:"required"`
}

type EnsureOrgRunnerGroupResponse struct {
	RunnerID string `json:"runner_id"`
}

// EnsureOrgRunnerGroup returns the org's runner ID, creating the org runner
// group first if it is missing. It is used by provision/reprovision to restore
// the org runner for orgs whose org-runner feature was re-enabled. Org signals
// are processed serially per org, so the check-then-create here is not subject
// to concurrent reprovisions.
//
// @temporal-gen-v2 activity
// @by-field OrgID
func (a *Activities) EnsureOrgRunnerGroup(ctx context.Context, req EnsureOrgRunnerGroupRequest) (*EnsureOrgRunnerGroupResponse, error) {
	org, err := a.getOrg(ctx, req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("unable to get org: %w", err)
	}

	if enabled, ok := org.Features[string(app.OrgFeatureOrgRunner)]; ok && !enabled {
		return nil, fmt.Errorf("org-runner feature disabled; refusing to create org runner group")
	}

	if len(org.RunnerGroup.Runners) > 0 {
		return &EnsureOrgRunnerGroupResponse{RunnerID: org.RunnerGroup.Runners[0].ID}, nil
	}

	rg, err := a.runnersHelpers.CreateOrgRunnerGroup(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("unable to create org runner group: %w", err)
	}
	if len(rg.Runners) == 0 {
		return nil, fmt.Errorf("created org runner group has no runners")
	}

	return &EnsureOrgRunnerGroupResponse{RunnerID: rg.Runners[0].ID}, nil
}
