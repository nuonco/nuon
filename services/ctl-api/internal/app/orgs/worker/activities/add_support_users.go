package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
)

type AddSupportUsersRequest struct {
	OrgID string `json:"org_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field OrgID
func (a *Activities) AddSupportUsers(ctx context.Context, req AddSupportUsersRequest) ([]helpers.SupportUserResult, error) {
	org, err := a.getOrg(ctx, req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("unable to get org: %w", err)
	}

	results, err := a.helpers.AddSupportUsersToOrg(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("unable to add support users: %w", err)
	}

	return results, nil
}
