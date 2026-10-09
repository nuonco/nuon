package authz

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/analytics/events"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (h *Client) AcceptInvite(ctx context.Context, invite *app.OrgInvite, acct *app.Account) error {
	roleType := invite.RoleType
	if roleType == "" {
		roleType = app.RoleTypeOrgAdmin
	}

	if err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := RequireUserManaged(tx, acct.ID); err != nil {
			return err
		}
		if err := New(Params{DB: tx}).AddAccountOrgRole(ctx, roleType, invite.OrgID, acct.ID); err != nil {
			return fmt.Errorf("unable to add account role: %w", err)
		}
		res := tx.Model(&app.OrgInvite{ID: invite.ID}).Updates(app.OrgInvite{Status: app.OrgInviteStatusAccepted})
		if res.Error != nil {
			return fmt.Errorf("unable to update invite: %w", res.Error)
		}
		if res.RowsAffected < 1 {
			return fmt.Errorf("invite not found %w", gorm.ErrRecordNotFound)
		}
		return nil
	}); err != nil {
		return err
	}

	// send a notification to the correct org event flow that it was accepted
	cctx.SetOrgContext(ctx, &invite.Org)

	h.analyticsClient.Track(ctx, events.InviteAccepted, map[string]interface{}{
		"invite_id": invite.ID,
		"email":     invite.Email,
		"org_id":    invite.OrgID,
		"role_type": invite.RoleType,
	})

	// return nil
	return nil
}
