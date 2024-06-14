package invites

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (m *middleware) handleInvites(ctx context.Context, acct *app.Account) error {
	var invites []app.OrgInvite
	res := m.db.
		WithContext(ctx).
		Where(&app.OrgInvite{
			Email:  acct.Email,
			Status: app.OrgInviteStatusPending,
		}).
		Find(&invites)
	if res.Error != nil {
		return res.Error
	}
	if len(invites) < 1 {
		return nil
	}

	for _, invite := range invites {
		if err := m.authz.AcceptInvite(ctx, &invite, acct); err != nil {
			return nil
		}
	}

	return nil
}
