package authz

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// RemoveAccountOrgRoleByType drops one role binding, unlike RemoveAccountOrgRoles,
// which clears every binding in an org.
func (h *Client) RemoveAccountOrgRoleByType(ctx context.Context, roleType app.RoleType, orgID, accountID string) error {
	var roleIDs []string
	if res := h.db.WithContext(ctx).
		Model(&app.Role{}).
		Where(app.Role{OrgID: generics.NewNullString(orgID), RoleType: roleType}).
		Pluck("id", &roleIDs); res.Error != nil {
		return errors.Wrap(res.Error, "unable to look up role")
	}

	if len(roleIDs) == 0 {
		return nil
	}

	if res := h.db.WithContext(ctx).
		Unscoped().
		Where("account_id = ? AND role_id IN ?", accountID, roleIDs).
		Delete(&app.AccountRole{}); res.Error != nil {
		return errors.Wrap(res.Error, "unable to remove role for account")
	}

	return nil
}
