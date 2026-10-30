package authz

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

// SetAccountOrgRole replaces an account's role(s) in an org with a single role
// of the given type. Unlike RemoveAccountOrgRoles it does not touch the
// account's OrgInvite records, so it is safe for in-place role changes.
func (h *Client) SetAccountOrgRole(ctx context.Context, orgID, accountID string, roleType app.RoleType) error {
	if orgID == "" || roleType == "" {
		return fmt.Errorf("org ID and role type are required")
	}
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockRoleOrg(tx, orgID); err != nil {
			return err
		}
		if err := lockLegacyRoleSubject(tx, accountID); err != nil {
			return err
		}
		var role app.Role
		if err := tx.
			Preload("Policies").
			Where(app.Role{
				OrgID:    generics.NewNullString(orgID),
				RoleType: roleType,
			}).
			Where(tx.Where(app.Role{Version: app.RoleVersionV1}).Or(app.Role{Version: app.RoleVersionV2})).
			First(&role).Error; err != nil {
			return fmt.Errorf("unable to find role: %w", err)
		}

		role, err := legacyMutationRole(tx, accountID, role)
		if err != nil {
			return err
		}
		if err := tx.
			Unscoped().
			Where(app.AccountRole{
				OrgID:     generics.NewNullString(orgID),
				AccountID: accountID,
			}).
			Delete(&app.AccountRole{}).Error; err != nil {
			return fmt.Errorf("unable to remove existing roles for account: %w", err)
		}

		if err := assertRoleAPICompatible(tx, accountID, &role); err != nil {
			return err
		}
		acctRole := &app.AccountRole{
			OrgID:     generics.NewNullString(orgID),
			RoleID:    role.ID,
			AccountID: accountID,
		}
		if role.Version == app.RoleVersionV3 {
			acctRole.ScopeKind = permissions.KindOrg
			acctRole.ScopeID = orgID
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&acctRole).Error; err != nil {
			return fmt.Errorf("unable to add role for account: %w", err)
		}

		if role.Version == app.RoleVersionV3 {
			_, err = loadScopedGrants(tx, accountID, orgID)
			return err
		}
		return nil
	})
}
