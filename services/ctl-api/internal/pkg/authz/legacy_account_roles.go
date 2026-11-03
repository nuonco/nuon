package authz

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func lockRoleOrg(tx *gorm.DB, orgID string) error {
	_, err := resolveTarget(tx.Clauses(clause.Locking{Strength: "UPDATE"}).Session(&gorm.Session{}), permissions.Scope{
		Kind: permissions.KindOrg, OrgID: orgID, ID: orgID,
	})
	return err
}

func lockLegacyRoleSubject(tx *gorm.DB, accountID string) error {
	if _, err := lockRoleSubject(tx, accountID); err != nil {
		return err
	}
	var bindings []app.AccountRole
	if err := tx.Preload("Role").Where(app.AccountRole{AccountID: accountID}).Find(&bindings).Error; err != nil {
		return err
	}
	for _, binding := range bindings {
		if binding.ScopeKind != "" || binding.ScopeID != "" || binding.Role.Version < app.RoleVersionV1 || binding.Role.Version > app.RoleVersionV2 {
			return fmt.Errorf("legacy role mutation cannot change scoped or unsupported bindings")
		}
	}
	return nil
}
