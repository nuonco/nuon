package migrations

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (m *Migrations) migration052CreateOrgPermissions(ctx context.Context) error {
	var userOrgs []app.UserOrg_deprecated
	res := m.db.WithContext(ctx).
		Find(&userOrgs)

	if res.Error != nil {
		return res.Error
	}

	for _, userOrg := range userOrgs {
		acct, err := m.authzClient.FetchAccount(ctx, userOrg.UserID)
		if err != nil {
			m.l.Info("skipping")
			continue
		}

		if err := m.authzClient.AddAccountRole(ctx, app.RoleTypeOrgAdmin, userOrg.OrgID, acct.ID); err != nil {
			return err
		}

		m.l.Info("success")
	}

	return nil
}
