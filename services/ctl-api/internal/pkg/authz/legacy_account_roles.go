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
	if err := tx.Preload("Role.Policies").Where(app.AccountRole{AccountID: accountID}).Find(&bindings).Error; err != nil {
		return err
	}
	for _, binding := range bindings {
		if binding.Role.Version == app.RoleVersionV3 {
			if err := binding.ValidateLegacyCompatibility(); err != nil {
				return err
			}
			continue
		}
		if binding.ScopeKind != "" || binding.ScopeID != "" || binding.Role.Version < app.RoleVersionV1 || binding.Role.Version > app.RoleVersionV2 {
			return fmt.Errorf("legacy role mutation cannot change scoped or unsupported bindings")
		}
	}
	return nil
}

func legacyMutationRole(tx *gorm.DB, accountID string, role app.Role) (app.Role, error) {
	if role.OrgID.Empty() {
		return role, nil
	}
	var bindings []app.AccountRole
	if err := tx.Preload("Role.Policies").Where(app.AccountRole{AccountID: accountID, OrgID: role.OrgID}).Find(&bindings).Error; err != nil {
		return app.Role{}, err
	}
	converted := false
	for _, binding := range bindings {
		if binding.Role.Version == app.RoleVersionV3 {
			if err := binding.ValidateLegacyCompatibility(); err != nil {
				return app.Role{}, err
			}
			converted = true
		}
	}
	if !converted || role.Version == app.RoleVersionV3 {
		return role, nil
	}
	if err := validateLegacyCustomerRole(role, role.OrgID.ValueString()); err != nil {
		return app.Role{}, err
	}
	return ensureConvertedCustomerRole(tx, role.OrgID.ValueString(), role.RoleType, role.Version)
}
