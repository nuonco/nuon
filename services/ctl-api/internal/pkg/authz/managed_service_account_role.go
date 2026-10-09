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
	if want == nil || want.OrgID.Empty() || want.RoleType == "" || want.Managed || len(want.Contexts) != 0 || len(want.Policies) != 1 || len(want.Policies[0].Permissions) == 0 {
		return fmt.Errorf("managed service account requires a private role with one explicit policy")
	}
	orgID := want.OrgID.ValueString()
	for object, verb := range want.Policies[0].Permissions {
		if object != orgID && !strings.HasPrefix(object, orgID+":") {
			return fmt.Errorf("private role grant %q is outside the owning org", object)
		}
		if verb == nil {
			return fmt.Errorf("private role grant requires a permission")
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
	if want.ID == "" {
		want.Policies[0].OrgID = want.OrgID
		if err := tx.Create(want).Error; err != nil {
			return fmt.Errorf("create private service account role: %w", err)
		}
	} else {
		var role app.Role
		if err := tx.Preload("Policies").Where(app.Role{ID: want.ID}).First(&role).Error; err != nil {
			return fmt.Errorf("load private service account role: %w", err)
		}
		if err := validatePrivateRole(tx, accountID, want.OrgID.ValueString(), &role); err != nil {
			return err
		}
		if role.RoleType != want.RoleType {
			return fmt.Errorf("private service account role has a different role type")
		}
		if len(role.Policies) == 0 {
			policy := want.Policies[0]
			policy.RoleID = want.ID
			policy.OrgID = want.OrgID
			if err := tx.Create(&policy).Error; err != nil {
				return fmt.Errorf("create private service account policy: %w", err)
			}
		} else if len(role.Policies) != 1 {
			return fmt.Errorf("private service account role has multiple policies")
		} else if !reflect.DeepEqual(role.Policies[0].Permissions, want.Policies[0].Permissions) {
			if err := tx.Model(&app.Policy{}).Where(app.Policy{ID: role.Policies[0].ID}).
				Update("permissions", want.Policies[0].Permissions).Error; err != nil {
				return fmt.Errorf("converge private service account policy: %w", err)
			}
		}
	}

	var count int64
	if err := tx.Model(&app.AccountRole{}).Where(app.AccountRole{RoleID: want.ID, AccountID: accountID}).Count(&count).Error; err != nil {
		return fmt.Errorf("load private role assignment: %w", err)
	}
	if count == 0 {
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

func DeleteManagedServiceAccountRole(tx *gorm.DB, accountID string) error {
	var binding app.ManagedServiceAccount
	if err := tx.Unscoped().Where(app.ManagedServiceAccount{AccountID: accountID}).Find(&binding).Error; err != nil {
		return fmt.Errorf("load private role ownership for deletion: %w", err)
	}
	if binding.PrivateRoleID == nil {
		return nil
	}
	var role app.Role
	if err := tx.Unscoped().Where(app.Role{ID: *binding.PrivateRoleID}).First(&role).Error; err != nil {
		return fmt.Errorf("load private service account role for deletion: %w", err)
	}
	if err := validatePrivateRole(tx, accountID, binding.OrgID, &role); err != nil {
		return err
	}
	if err := tx.Unscoped().Model(&binding).Update("private_role_id", nil).Error; err != nil {
		return fmt.Errorf("clear private role ownership: %w", err)
	}
	return deletePrivateRole(tx, accountID, role.ID)
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
