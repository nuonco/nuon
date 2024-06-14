package authz

import (
	"context"
	"fmt"

	"gorm.io/gorm/clause"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (h *Client) AddAccountRole(ctx context.Context, roleType app.RoleType, orgID, accountID string) error {
	var role app.Role
	res := h.db.WithContext(ctx).
		Where(app.Role{
			OrgID:    orgID,
			RoleType: roleType,
		}).
		First(&role)
	if res.Error != nil {
		return fmt.Errorf("unable to find role: %w", res.Error)
	}

	acctRole := &app.AccountRole{
		OrgID:     orgID,
		RoleID:    role.ID,
		AccountID: accountID,
	}

	res = h.db.WithContext(ctx).Clauses(
		clause.OnConflict{DoNothing: true},
	).Create(&acctRole)
	if res.Error != nil {
		return fmt.Errorf("unable to add role for account: %w", res.Error)
	}

	return nil
}
