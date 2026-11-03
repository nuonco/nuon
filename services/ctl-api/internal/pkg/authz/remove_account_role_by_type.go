package authz

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// RemoveAccountOrgRoleByType drops one role binding, unlike RemoveAccountOrgRoles,
// which clears every binding in an org.
func (h *Client) RemoveAccountOrgRoleByType(ctx context.Context, roleType app.RoleType, orgID, accountID string) error {
	if orgID == "" || roleType == "" {
		return errors.New("org ID and role type are required")
	}
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockLegacyRoleSubject(tx, accountID); err != nil {
			return err
		}
		var roleIDs []string
		if res := tx.Model(&app.Role{}).
			Where(app.Role{OrgID: generics.NewNullString(orgID), RoleType: roleType}).
			Pluck("id", &roleIDs); res.Error != nil {
			return errors.Wrap(res.Error, "unable to look up role")
		}
		if len(roleIDs) == 0 {
			return nil
		}
		if res := tx.Unscoped().
			Where(app.AccountRole{AccountID: accountID, OrgID: generics.NewNullString(orgID)}).
			Where("role_id IN ?", roleIDs).
			Delete(&app.AccountRole{}); res.Error != nil {
			return errors.Wrap(res.Error, "unable to remove role for account")
		}
		return nil
	})
}
