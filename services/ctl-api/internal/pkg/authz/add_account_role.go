package authz

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (h *Client) AddAccountOrgRole(ctx context.Context, roleType app.RoleType, orgID, accountID string) error {
	if orgID == "" || roleType == "" {
		return fmt.Errorf("org ID and role type are required")
	}
	return h.addLegacyAccountRole(ctx, accountID, app.Role{OrgID: generics.NewNullString(orgID), RoleType: roleType})
}

func (h *Client) AddAccountRoleByID(ctx context.Context, roleID, accountID string) error {
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}
	return h.addLegacyAccountRole(ctx, accountID, app.Role{ID: roleID})
}

func (h *Client) addLegacyAccountRole(ctx context.Context, accountID string, filter app.Role) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockLegacyRoleSubject(tx, accountID); err != nil {
			return err
		}
		var role app.Role
		if err := tx.Preload("Policies").Where(filter).First(&role).Error; err != nil {
			return fmt.Errorf("unable to find role: %w", err)
		}
		if err := assertRoleAPICompatible(tx, accountID, &role); err != nil {
			return err
		}
		binding := app.AccountRole{OrgID: role.OrgID, RoleID: role.ID, AccountID: accountID}
		var count int64
		if err := tx.Model(&app.AccountRole{}).Where(binding).Count(&count).Error; err != nil {
			return err
		}
		if count != 0 {
			return nil
		}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&binding).Error
	})
}
