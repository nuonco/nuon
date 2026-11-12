package authz

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (h *Client) EnsureManagedServiceAccountRole(ctx context.Context, accountID string, want *app.Role) error {
	if want == nil || want.OrgID.Empty() || want.RoleType == "" || want.Managed || len(want.Contexts) != 0 || len(want.Policies) != 1 {
		return fmt.Errorf("managed service account requires a private role with one explicit policy")
	}
	orgID := want.OrgID.ValueString()
	want.Policies[0].OrgID = want.OrgID
	projection, err := want.Policies[0].PermissionProjection()
	if err != nil {
		return err
	}
	if len(projection) == 0 {
		return fmt.Errorf("managed service account requires an explicit grant")
	}
	for object, verb := range projection {
		if verb == nil {
			return fmt.Errorf("private role grant requires a permission")
		}
		if object == permissions.APIPermissionKey {
			if _, err := permissions.ParseAPI(*verb); err != nil {
				return err
			}
			continue
		}
		if object != orgID && !strings.HasPrefix(object, orgID+":") {
			return fmt.Errorf("private role grant %q is outside the owning org", object)
		}
		if _, err := permissions.NewPermission(*verb); err != nil {
			return err
		}
	}

	ctx = cctx.SetAccountIDContext(ctx, accountID)
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var acct app.Account
		if err := tx.Select("id", "account_type").Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(app.Account{ID: accountID}).Take(&acct).Error; err != nil {
			return fmt.Errorf("lock managed service account for role reconciliation: %w", err)
		}
		if acct.AccountType != app.AccountTypeService {
			return fmt.Errorf("private roles require a service account")
		}
		var binding app.ManagedServiceAccount
		if err := tx.Where(app.ManagedServiceAccount{AccountID: accountID, OrgID: orgID}).First(&binding).Error; err != nil {
			return fmt.Errorf("load private role ownership: %w", err)
		}

		roleID := want.ID
		if binding.PrivateRoleID != nil {
			if roleID != "" && roleID != *binding.PrivateRoleID {
				return fmt.Errorf("managed account already owns a different private role")
			}
			roleID = *binding.PrivateRoleID
		}
		want.ID = roleID
		if err := ensurePrivateRole(tx, accountID, want); err != nil {
			return err
		}
		if binding.PrivateRoleID == nil {
			if err := tx.Model(&binding).Update("private_role_id", want.ID).Error; err != nil {
				return fmt.Errorf("record private service account role ownership: %w", err)
			}
		}
		return nil
	})
}

func ensurePrivateRole(tx *gorm.DB, accountID string, want *app.Role) error {
	if !want.RoleType.IsPrivate() {
		return fmt.Errorf("role type %q is not a private role type", want.RoleType)
	}
	if want.ID == "" {
		want.Policies[0].OrgID = want.OrgID
		if err := tx.Create(want).Error; err != nil {
			return fmt.Errorf("create private service account role: %w", err)
		}
	} else {
		var role app.Role
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Policies").Where(app.Role{ID: want.ID}).First(&role).Error; err != nil {
			return fmt.Errorf("load private service account role: %w", err)
		}
		if err := validatePrivateRole(tx, accountID, want.OrgID.ValueString(), &role); err != nil {
			return err
		}
		if role.RoleType != want.RoleType {
			return fmt.Errorf("private service account role has a different role type")
		}
		if len(role.Policies) == 0 {
			if role.Version >= app.RoleVersionV3 {
				return fmt.Errorf("private service account role is missing its document policy")
			}
			policy := want.Policies[0]
			policy.RoleID = want.ID
			policy.OrgID = want.OrgID
			if err := tx.Create(&policy).Error; err != nil {
				return fmt.Errorf("create private service account policy: %w", err)
			}
			role.Policies = []app.Policy{policy}
		} else if len(role.Policies) != 1 {
			return fmt.Errorf("private service account role has multiple policies")
		} else {
			existing := &role.Policies[0]
			current, err := existing.PermissionProjection()
			if err != nil {
				return err
			}
			desired, err := want.Policies[0].PermissionProjection()
			if err != nil {
				return err
			}
			var document []byte
			if existing.Document != nil {
				delete(desired, permissions.APIPermissionKey)
				version := app.RoleVersionV1
				if api, ok := current[permissions.APIPermissionKey]; ok {
					desired[permissions.APIPermissionKey] = api
					version = app.RoleVersionV2
				}
				target := app.Role{OrgID: role.OrgID, RoleType: role.RoleType, Version: version,
					Policies: []app.Policy{{OrgID: role.OrgID, Permissions: desired}}}
				policy, err := ScopedRolePolicy(&target)
				if err != nil {
					return err
				}
				document, err = policy.MarshalJSON()
				if err != nil {
					return err
				}
			}
			if !reflect.DeepEqual(current, desired) {
				values := map[string]any{"permissions": desired, "document": nil}
				if existing.Document != nil {
					values = map[string]any{"permissions": nil, "document": document}
				}
				if err := tx.Model(&app.Policy{}).Where(app.Policy{ID: existing.ID}).Updates(values).Error; err != nil {
					return fmt.Errorf("converge private service account policy: %w", err)
				}
				if existing.Document != nil {
					existing.Document, existing.Permissions = document, nil
				} else {
					existing.Permissions = desired
				}
			}
		}
		want.Version, want.Policies = role.Version, role.Policies
	}

	var count int64
	if err := tx.Model(&app.AccountRole{}).Where(app.AccountRole{RoleID: want.ID, AccountID: accountID}).Count(&count).Error; err != nil {
		return fmt.Errorf("load private role assignment: %w", err)
	}
	if count == 0 {
		if err := assertRoleAPICompatible(tx, accountID, want); err != nil {
			return err
		}
		if err := tx.Create(&app.AccountRole{
			OrgID: generics.NewNullString(want.OrgID.ValueString()), RoleID: want.ID, AccountID: accountID,
		}).Error; err != nil {
			return fmt.Errorf("assign private service account role: %w", err)
		}
	}
	return nil
}

func validatePrivateRole(tx *gorm.DB, accountID, orgID string, role *app.Role) error {
	if role.OrgID.ValueString() != orgID || role.Managed || len(role.Contexts) != 0 {
		return fmt.Errorf("role %s is not a private role for the owning org", role.ID)
	}
	var count int64
	if err := tx.Unscoped().Model(&app.AccountRole{}).
		Where(app.AccountRole{RoleID: role.ID}).Where("account_id <> ?", accountID).Count(&count).Error; err != nil {
		return fmt.Errorf("validate private role assignments: %w", err)
	}
	if count != 0 {
		return fmt.Errorf("role %s is also assigned to another account", role.ID)
	}
	return nil
}

func DeleteManagedServiceAccountRoles(tx *gorm.DB, accountIDs []string) error {
	var bindings []app.ManagedServiceAccount
	if err := tx.Unscoped().Where("account_id IN ? AND private_role_id IS NOT NULL", accountIDs).Find(&bindings).Error; err != nil {
		return fmt.Errorf("load private role ownership for deletion: %w", err)
	}
	for _, binding := range bindings {
		var role app.Role
		if err := tx.Unscoped().Where(app.Role{ID: *binding.PrivateRoleID}).First(&role).Error; err != nil {
			return fmt.Errorf("load private service account role for deletion: %w", err)
		}
		if err := validatePrivateRole(tx, binding.AccountID, binding.OrgID, &role); err != nil {
			return err
		}
		if err := tx.Unscoped().Model(&binding).Update("private_role_id", nil).Error; err != nil {
			return fmt.Errorf("clear private role ownership: %w", err)
		}
		if err := deletePrivateRole(tx, binding.AccountID, role.ID); err != nil {
			return err
		}
	}
	return nil
}

func DeletePrivateRoles(tx *gorm.DB, accountIDs []string) error {
	roleIDs, blocked, err := privateRolesOf(tx, accountIDs)
	if err != nil {
		return err
	}
	for accountID, roleID := range blocked {
		return fmt.Errorf("role %s held by account %s is shared or not a private role", roleID, accountID)
	}
	if len(roleIDs) == 0 {
		return nil
	}
	if err := tx.Unscoped().Where("account_id IN ? AND role_id IN ?", accountIDs, roleIDs).Delete(&app.AccountRole{}).Error; err != nil {
		return fmt.Errorf("remove private role assignments: %w", err)
	}
	if err := tx.Unscoped().Where("role_id IN ?", roleIDs).Delete(&app.Policy{}).Error; err != nil {
		return fmt.Errorf("delete private role policies: %w", err)
	}
	if err := tx.Unscoped().Where("id IN ?", roleIDs).Delete(&app.Role{}).Error; err != nil {
		return fmt.Errorf("delete private roles: %w", err)
	}
	return nil
}

func PrivateRoleBlockedAccounts(tx *gorm.DB, accountIDs []string) (map[string]struct{}, error) {
	_, blocked, err := privateRolesOf(tx, accountIDs)
	if err != nil {
		return nil, err
	}
	out := make(map[string]struct{}, len(blocked))
	for accountID := range blocked {
		out[accountID] = struct{}{}
	}
	return out, nil
}

func privateRolesOf(tx *gorm.DB, accountIDs []string) ([]string, map[string]string, error) {
	if len(accountIDs) == 0 {
		return nil, nil, nil
	}
	var holders []app.AccountRole
	if err := tx.Unscoped().Model(&app.AccountRole{}).
		Select("account_roles.account_id", "account_roles.role_id").
		Joins("JOIN roles ON roles.id = account_roles.role_id").
		Where("account_roles.account_id IN ?", accountIDs).
		Where("roles.role_type IN ?", app.PrivateRoleTypes).
		Find(&holders).Error; err != nil {
		return nil, nil, fmt.Errorf("load private roles for deletion: %w", err)
	}
	if len(holders) == 0 {
		return nil, nil, nil
	}
	roleIDs := make([]string, 0, len(holders))
	seen := make(map[string]struct{}, len(holders))
	for _, h := range holders {
		if _, ok := seen[h.RoleID]; !ok {
			seen[h.RoleID] = struct{}{}
			roleIDs = append(roleIDs, h.RoleID)
		}
	}

	bad := map[string]struct{}{}
	var roles []app.Role
	if err := tx.Unscoped().Select("id", "managed", "contexts").Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		return nil, nil, fmt.Errorf("load private roles for validation: %w", err)
	}
	for _, role := range roles {
		if role.Managed || len(role.Contexts) != 0 {
			bad[role.ID] = struct{}{}
		}
	}
	var shared []string
	if err := tx.Unscoped().Model(&app.AccountRole{}).
		Where("role_id IN ? AND account_id NOT IN ?", roleIDs, accountIDs).
		Distinct("role_id").Pluck("role_id", &shared).Error; err != nil {
		return nil, nil, fmt.Errorf("validate private role assignments: %w", err)
	}
	for _, roleID := range shared {
		bad[roleID] = struct{}{}
	}

	blocked := map[string]string{}
	for changed := true; changed; {
		changed = false
		for _, h := range holders {
			if _, ok := bad[h.RoleID]; ok {
				if _, ok := blocked[h.AccountID]; !ok {
					blocked[h.AccountID] = h.RoleID
					changed = true
				}
			}
		}
		for _, h := range holders {
			if _, ok := blocked[h.AccountID]; ok {
				if _, ok := bad[h.RoleID]; !ok {
					bad[h.RoleID] = struct{}{}
					changed = true
				}
			}
		}
	}
	return roleIDs, blocked, nil
}

func deletePrivateRole(tx *gorm.DB, accountID, roleID string) error {
	if err := tx.Unscoped().Where(app.AccountRole{AccountID: accountID, RoleID: roleID}).Delete(&app.AccountRole{}).Error; err != nil {
		return fmt.Errorf("remove private role assignments: %w", err)
	}
	if err := tx.Unscoped().Where(app.Policy{RoleID: roleID}).Delete(&app.Policy{}).Error; err != nil {
		return fmt.Errorf("delete private role policy: %w", err)
	}
	if err := tx.Unscoped().Delete(&app.Role{ID: roleID}).Error; err != nil {
		return fmt.Errorf("delete private role: %w", err)
	}
	return nil
}
