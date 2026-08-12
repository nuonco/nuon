package authz

import (
	"context"
	"fmt"

	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// AddAccountOrgRole grants the account a role named by role type — or by role
// id for custom roles, which all share app.RoleTypeCustom.
func (h *Client) AddAccountOrgRole(ctx context.Context, roleType app.RoleType, orgID, accountID string) error {
	role, err := findOrgRole(h.db.WithContext(ctx), orgID, roleType)
	if err != nil {
		return err
	}

	acctRole := &app.AccountRole{
		OrgID:     generics.NewNullString(orgID),
		RoleID:    role.ID,
		AccountID: accountID,
	}

	res := h.db.WithContext(ctx).Clauses(
		clause.OnConflict{DoNothing: true},
	).Create(&acctRole)
	if res.Error != nil {
		return fmt.Errorf("unable to add role for account: %w", res.Error)
	}

	return nil
}

func (h *Client) AddAccountRoleByID(ctx context.Context, roleID, accountID string) error {
	acctRole := &app.AccountRole{
		RoleID:    roleID,
		AccountID: accountID,
	}

	res := h.db.WithContext(ctx).Clauses(
		clause.OnConflict{DoNothing: true},
	).Create(&acctRole)
	if res.Error != nil {
		return fmt.Errorf("unable to add role for account: %w", res.Error)
	}

	return nil
}
