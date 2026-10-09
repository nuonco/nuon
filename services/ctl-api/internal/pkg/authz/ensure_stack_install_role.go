package authz

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

// stackInstallRole is the role an install stack's account holds: every stack
// operation on its own install and nothing else. The grant is `all` because the key
// is already scoped to one install's stack namespace.
func stackInstallRole(orgID, installID string) *app.Role {
	return &app.Role{
		OrgID:       generics.NewNullString(orgID),
		RoleType:    app.RoleTypeStack,
		Title:       "Stack",
		Description: "Scoped access to install-stack endpoints for a single install.",
		Policies: []app.Policy{{
			OrgID: generics.NewNullString(orgID),
			Name:  app.PolicyNameStack,
			Permissions: pgtype.Hstore{
				permissions.StackObject(orgID, installID): permissions.PermissionAll.ToStrPtr(),
			},
		}},
	}
}

// EnsureStackInstallRole converges a stack account's install-scoped role: one role,
// one policy, on this install's stack object. Looked up through the account's own
// binding, since the account is already per-install.
func (h *Client) EnsureStackInstallRole(ctx context.Context, orgID, installID, accountID string) error {
	ctx = cctx.SetAccountIDContext(ctx, accountID)
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var acct app.Account
		if err := tx.Select("id", "account_type", "subject").Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(app.Account{ID: accountID}).Take(&acct).Error; err != nil {
			return fmt.Errorf("lock stack service account: %w", err)
		}
		if acct.AccountType != app.AccountTypeService {
			return fmt.Errorf("stack roles require a service account")
		}
		var binding app.ManagedServiceAccount
		if err := tx.Where(app.ManagedServiceAccount{AccountID: accountID}).Find(&binding).Error; err != nil {
			return fmt.Errorf("load stack service account ownership: %w", err)
		}
		if binding.AccountID != "" {
			if binding.OrgID != orgID || binding.OwnerType != plugins.TableName(tx, app.InstallStack{}) || binding.Purpose != app.ManagedServiceAccountPurposeStack || binding.OwnerID != acct.Subject {
				return fmt.Errorf("service account is not owned by this stack")
			}
			var stack struct{ ID string }
			if err := tx.Model(&app.InstallStack{}).Select("id").
				Where(app.InstallStack{ID: binding.OwnerID, OrgID: orgID, InstallID: installID}).Take(&stack).Error; err != nil {
				return fmt.Errorf("validate stack permission target: %w", err)
			}
		}

		want := stackInstallRole(orgID, installID)
		if binding.PrivateRoleID == nil {
			var existing []app.Role
			if err := tx.Joins("JOIN account_roles ON account_roles.role_id = roles.id AND account_roles.deleted_at = 0").
				Preload("Policies").Where("account_roles.account_id = ?", accountID).
				Where(app.Role{OrgID: generics.NewNullString(orgID), RoleType: app.RoleTypeStack}).Find(&existing).Error; err != nil {
				return fmt.Errorf("load legacy stack roles: %w", err)
			}
			if len(existing) > 1 {
				return fmt.Errorf("stack service account has multiple legacy stack roles")
			}
			if len(existing) == 1 {
				role := existing[0]
				if len(role.Policies) != 1 || len(role.Policies[0].Permissions) != 1 {
					return fmt.Errorf("legacy stack role has an unexpected policy")
				}
				if _, ok := role.Policies[0].Permissions[permissions.StackObject(orgID, installID)]; !ok {
					return fmt.Errorf("legacy stack role targets a different resource")
				}
				want.ID = role.ID
			}
		}
		if binding.AccountID != "" {
			return New(Params{DB: tx}).EnsureManagedServiceAccountRole(ctx, accountID, want)
		}
		return ensurePrivateRole(tx, accountID, want)
	})
}

func DeleteStackInstallRoles(tx *gorm.DB, accountID string) error {
	var roles []app.Role
	if err := tx.Unscoped().Joins("JOIN account_roles ON account_roles.role_id = roles.id").
		Where("account_roles.account_id = ?", accountID).Where(app.Role{RoleType: app.RoleTypeStack}).
		Distinct("roles.*").Find(&roles).Error; err != nil {
		return fmt.Errorf("load legacy stack roles for deletion: %w", err)
	}
	for _, role := range roles {
		if err := validatePrivateRole(tx, accountID, role.OrgID.ValueString(), &role); err != nil {
			return err
		}
		if err := deletePrivateRole(tx, accountID, role.ID); err != nil {
			return err
		}
	}
	return nil
}
