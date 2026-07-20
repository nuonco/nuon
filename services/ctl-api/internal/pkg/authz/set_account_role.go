package authz

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// SetAccountOrgRole replaces an account's role(s) in an org with a single role
// of the given type. Unlike RemoveAccountOrgRoles it does not touch the
// account's OrgInvite records, so it is safe for in-place role changes.
func (h *Client) SetAccountOrgRole(ctx context.Context, orgID, accountID string, roleType app.RoleType) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var role app.Role
		if err := tx.
			Where(app.Role{
				OrgID:    generics.NewNullString(orgID),
				RoleType: roleType,
			}).
			First(&role).Error; err != nil {
			return fmt.Errorf("unable to find role: %w", err)
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

		acctRole := &app.AccountRole{
			OrgID:     generics.NewNullString(orgID),
			RoleID:    role.ID,
			AccountID: accountID,
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&acctRole).Error; err != nil {
			return fmt.Errorf("unable to add role for account: %w", err)
		}

		return nil
	})
}
