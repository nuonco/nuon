package authz

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/analytics/events"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

func (h *Client) AcceptInvite(ctx context.Context, invite *app.OrgInvite, acct *app.Account) error {
	// add the role to the user
	if err := h.AddAccountOrgRole(ctx, app.RoleTypeOrgAdmin, invite.OrgID, acct.ID); err != nil {
		return fmt.Errorf("unable to add account role: %w", err)
	}

	// update invite object
	res := h.db.WithContext(ctx).
		Model(&app.OrgInvite{ID: invite.ID}).
		Updates(app.OrgInvite{Status: app.OrgInviteStatusAccepted})
	if res.Error != nil {
		return fmt.Errorf("unable to update invite: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("invite not found %w", gorm.ErrRecordNotFound)
	}

	// send a notification to the correct org event flow that it was accepted
	cctx.SetOrgContext(ctx, &invite.Org)

	h.evClient.Send(ctx, invite.OrgID, &signals.Signal{
		Type:     signals.OperationInviteAccepted,
		InviteID: invite.ID,
	})

	h.analyticsClient.Track(ctx, events.InviteAccepted, map[string]interface{}{
		"invite_id": invite.ID,
		"email":     invite.Email,
		"org_id":    invite.OrgID,
		"role_type": invite.RoleType,
	})

	// return nil
	return nil
}
