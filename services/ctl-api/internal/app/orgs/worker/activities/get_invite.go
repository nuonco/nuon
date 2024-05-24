package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetInviteRequest struct {
	InviteID string `validate:"required"`
}

func (a *Activities) GetInvite(ctx context.Context, req GetInviteRequest) (*app.OrgInvite, error) {
	org, err := a.getOrgInvite(ctx, req.InviteID)
	if err != nil {
		return nil, fmt.Errorf("unable to get org invite: %w", err)
	}

	return org, nil
}

func (a *Activities) getOrgInvite(ctx context.Context, inviteID string) (*app.OrgInvite, error) {
	orgInvite := app.OrgInvite{}
	res := a.db.WithContext(ctx).
		Preload("CreatedBy").
		First(&orgInvite, "id = ?", inviteID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org invite: %w", res.Error)
	}

	return &orgInvite, nil
}
