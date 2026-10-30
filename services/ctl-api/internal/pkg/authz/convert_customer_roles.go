package authz

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func (c *Client) ConvertLegacyCustomerRoles(ctx context.Context, accountID, orgID string) error {
	scope := permissions.Scope{Kind: permissions.KindOrg, OrgID: orgID, ID: orgID}
	if err := scope.ValidateShape(); err != nil {
		return err
	}
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := requireRoleMutationIsolation(tx); err != nil {
			return err
		}
		if _, err := resolveTarget(tx.Clauses(clause.Locking{Strength: "UPDATE"}).Session(&gorm.Session{}), scope); err != nil {
			return err
		}
		account, err := lockRoleSubject(tx, accountID)
		if err != nil {
			return err
		}
		switch account.AccountType {
		case app.AccountTypeAuth, app.AccountTypeAuth0, app.AccountTypeService:
		default:
			return fmt.Errorf("unsupported customer role subject type %q", account.AccountType)
		}
		var managed int64
		if err := tx.Unscoped().Model(&app.ManagedServiceAccount{}).Where(app.ManagedServiceAccount{AccountID: accountID}).Count(&managed).Error; err != nil {
			return err
		}
		if managed != 0 {
			return fmt.Errorf("managed account roles require owner-specific conversion")
		}
		var bindings []app.AccountRole
		if err := tx.Preload("Role", func(db *gorm.DB) *gorm.DB {
			return db.Clauses(clause.Locking{Strength: "SHARE"}).Order("id")
		}).Preload("Role.Policies", func(db *gorm.DB) *gorm.DB {
			return db.Clauses(clause.Locking{Strength: "SHARE"}).Order("id")
		}).Where(app.AccountRole{AccountID: accountID}).Find(&bindings).Error; err != nil {
			return err
		}
		var sources []string
		wanted := make(map[app.RoleType]int)
		for _, binding := range bindings {
			tenant := binding.OrgID.ValueString()
			if tenant == "" || binding.Role.OrgID.ValueString() != tenant || binding.Role.ID == "" || binding.RoleID != binding.Role.ID {
				return fmt.Errorf("customer binding requires a live role in its tenant")
			}
			if account.AccountType == app.AccountTypeService && tenant != orgID {
				return fmt.Errorf("service account bindings must remain in one tenant")
			}
			if binding.Role.Version == app.RoleVersionV3 {
				policy, err := decodeScopedRole(binding.Role, tenant)
				if err != nil {
					return err
				}
				if policy.Kind() != permissions.PolicyKindBundle || binding.Role.CatalogRevision <= 0 {
					return fmt.Errorf("customer conversion requires catalog bundles")
				}
				if _, err := policy.Compile(tenant, permissions.Scope{Kind: binding.ScopeKind, OrgID: tenant, ID: binding.ScopeID}); err != nil {
					return err
				}
				continue
			}
			if err := validateLegacyCustomerBinding(binding); err != nil {
				return err
			}
			if tenant == orgID {
				sources = append(sources, binding.ID)
				wanted[binding.Role.RoleType] = max(wanted[binding.Role.RoleType], binding.Role.Version)
			}
		}
		if len(sources) > 0 {
			if err := tx.Where(app.AccountRole{AccountID: accountID, OrgID: generics.NewNullString(orgID)}).
				Where("id IN ?", sources).Delete(&app.AccountRole{}).Error; err != nil {
				return err
			}
			for _, roleType := range []app.RoleType{app.RoleTypeOrgAdmin, app.RoleTypeOrgReadOnly} {
				if wanted[roleType] == 0 {
					continue
				}
				role, err := ensureConvertedCustomerRole(tx, orgID, roleType, wanted[roleType])
				if err != nil {
					return err
				}
				if err := addScopedAccountRole(tx, account, role.ID, scope); err != nil {
					return err
				}
			}
		}
		_, err = loadScopedGrants(tx, accountID, orgID)
		if len(sources) == 0 && errors.Is(err, errNoScopedGrants) {
			return nil
		}
		return err
	})
}

func validateLegacyCustomerBinding(binding app.AccountRole) error {
	if binding.ID == "" || binding.ScopeKind != "" || binding.ScopeID != "" {
		return fmt.Errorf("unsupported legacy customer binding shape")
	}
	return validateLegacyCustomerRole(binding.Role, binding.OrgID.ValueString())
}

func validateLegacyCustomerRole(role app.Role, orgID string) error {
	if err := (permissions.Scope{Kind: permissions.KindOrg, OrgID: orgID, ID: orgID}).ValidateShape(); err != nil {
		return err
	}
	if role.ID == "" || role.OrgID.ValueString() != orgID || role.CatalogRevision != 0 || role.LegacyVersion != 0 ||
		(role.Version != app.RoleVersionV1 && role.Version != app.RoleVersionV2) || len(role.Policies) != 1 {
		return fmt.Errorf("unsupported legacy customer role shape")
	}
	var verb permissions.Permission
	switch role.RoleType {
	case app.RoleTypeOrgAdmin:
		verb = permissions.PermissionAll
	case app.RoleTypeOrgReadOnly:
		verb = permissions.PermissionRead
	default:
		return fmt.Errorf("unsupported legacy customer role %q", role.RoleType)
	}
	policy := role.Policies[0]
	if policy.RoleID != role.ID || policy.OrgID.ValueString() != orgID || len(policy.Document) != 0 {
		return fmt.Errorf("legacy policy must belong to its role tenant and contain only legacy authority")
	}
	grant := policy.Permissions[orgID]
	if grant == nil || *grant != string(verb) {
		return fmt.Errorf("legacy customer role has an unexpected org grant")
	}
	wantKeys := 1
	if role.Version == app.RoleVersionV2 {
		api := policy.Permissions[permissions.APIPermissionKey]
		if api == nil || *api != string(permissions.APICtl) {
			return fmt.Errorf("legacy v2 customer role requires ctl-api eligibility")
		}
		wantKeys++
	}
	if len(policy.Permissions) != wantKeys {
		return fmt.Errorf("legacy customer role contains additional authority")
	}
	return nil
}

func ensureConvertedCustomerRole(tx *gorm.DB, orgID string, roleType app.RoleType, legacyVersion int) (app.Role, error) {
	definition, err := permissions.CurrentCustomerRole(permissions.CustomerRoleName(roleType))
	if err != nil {
		return app.Role{}, err
	}
	document, err := definition.MarshalPolicy()
	if err != nil {
		return app.Role{}, err
	}
	var existing []app.Role
	if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Preload("Policies", func(db *gorm.DB) *gorm.DB {
		return db.Clauses(clause.Locking{Strength: "SHARE"})
	}).Where(app.Role{OrgID: generics.NewNullString(orgID), RoleType: roleType, Version: app.RoleVersionV3, LegacyVersion: legacyVersion}).Find(&existing).Error; err != nil {
		return app.Role{}, err
	}
	if len(existing) > 1 {
		return app.Role{}, fmt.Errorf("multiple v3 catalog roles for customer role %q", roleType)
	}
	if len(existing) == 1 {
		role := existing[0]
		if err := role.ValidateLegacyCompatibility(); err != nil {
			return app.Role{}, err
		}
		return role, nil
	}
	for _, role := range standardOrgRoles(orgID) {
		if role.RoleType != roleType {
			continue
		}
		role.Version = app.RoleVersionV3
		role.LegacyVersion = legacyVersion
		role.CatalogRevision = definition.Revision
		if legacyVersion == app.RoleVersionV1 {
			delete(role.Policies[0].Permissions, permissions.APIPermissionKey)
		}
		role.Policies[0].Document = document
		if err := tx.Create(&role).Error; err != nil {
			return app.Role{}, err
		}
		return role, nil
	}
	return app.Role{}, fmt.Errorf("unsupported customer catalog role %q", roleType)
}
