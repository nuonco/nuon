package authz

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// findOrgRole resolves an org's role by role type, or by role id — the only
// way to address a custom role, since they all share app.RoleTypeCustom.
func findOrgRole(tx *gorm.DB, orgID string, identifier app.RoleType) (*app.Role, error) {
	if identifier == app.RoleTypeCustom {
		return nil, fmt.Errorf("custom roles must be addressed by role id")
	}

	var role app.Role
	if err := tx.
		Where(app.Role{OrgID: generics.NewNullString(orgID)}).
		Where(
			tx.Where(app.Role{RoleType: identifier}).
				Or(app.Role{ID: string(identifier)}),
		).
		First(&role).Error; err != nil {
		return nil, fmt.Errorf("unable to find role: %w", err)
	}

	return &role, nil
}
