package authz

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// stackInstallRole is the unmanaged, contextless role an install stack's service
// account holds: read on its own install's stack endpoints and nothing else.
func stackInstallRole(orgID, installID string) *app.Role {
	return &app.Role{
		OrgID:       generics.NewNullString(orgID),
		RoleType:    app.RoleTypeStack,
		Title:       "Stack",
		Description: "Scoped access to install-stack endpoints for a single install.",
		Managed:     false,
		Policies: []app.Policy{
			{
				OrgID: generics.NewNullString(orgID),
				Name:  app.PolicyNameStack,
				Permissions: pgtype.Hstore(map[string]*string{
					permissions.StackObject(orgID, installID): permissions.PermissionRead.ToStrPtr(),
				}),
			},
		},
	}
}

// EnsureStackInstallRole converges the install-scoped role for a stack service
// account: one role per account, holding one policy granting read on this
// install's stack object.
//
// The service account is already per-install, so its stack role is unambiguous
// and looked up through its own binding rather than by (org, role_type).
func (h *Client) EnsureStackInstallRole(ctx context.Context, orgID, installID, accountID string) error {
	// Role.BeforeCreate reads the creator from the account ID in the context, and
	// a temporal activity has none. The account is its own creator.
	ctx = cctx.SetAccountIDContext(ctx, accountID)

	var existing []app.Role
	if res := h.db.WithContext(ctx).
		Joins("JOIN account_roles ON account_roles.role_id = roles.id AND account_roles.deleted_at = 0").
		Preload("Policies").
		Where("account_roles.account_id = ?", accountID).
		Where("roles.org_id = ? AND roles.role_type = ?", orgID, app.RoleTypeStack).
		Find(&existing); res.Error != nil {
		return errors.Wrap(res.Error, "unable to look up stack install role")
	}

	want := permissions.StackObject(orgID, installID)
	for _, role := range existing {
		if len(role.Policies) != 1 {
			continue
		}
		if _, ok := role.Policies[0].Permissions[want]; ok {
			return nil
		}
	}

	role := stackInstallRole(orgID, installID)
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if res := tx.Create(role); res.Error != nil {
			return errors.Wrap(res.Error, "unable to create stack install role")
		}

		binding := &app.AccountRole{
			OrgID:     generics.NewNullString(orgID),
			RoleID:    role.ID,
			AccountID: accountID,
		}
		if res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(binding); res.Error != nil {
			return errors.Wrap(res.Error, "unable to bind stack install role")
		}

		return nil
	})
}

// DeleteStackInstallRoles hard-deletes the stack roles bound to an account, and
// their policies. They are per-account garbage once the account is gone, and a
// soft delete would keep the unique policy-per-role index occupied.
//
// tx is expected to be a transaction.
func DeleteStackInstallRoles(tx *gorm.DB, accountID string) error {
	var roleIDs []string
	if res := tx.Unscoped().
		Model(&app.Role{}).
		Joins("JOIN account_roles ON account_roles.role_id = roles.id").
		Where("account_roles.account_id = ?", accountID).
		Where("roles.role_type = ?", app.RoleTypeStack).
		Distinct().
		Pluck("roles.id", &roleIDs); res.Error != nil {
		return errors.Wrap(res.Error, "unable to look up stack roles for account")
	}

	if len(roleIDs) == 0 {
		return nil
	}

	// Bindings first: account_roles carries a foreign key to roles, so the role
	// rows cannot go until nothing points at them. Every binding on a per-install
	// stack role belongs to the account being deleted.
	if res := tx.Unscoped().
		Where("role_id IN ?", roleIDs).
		Delete(&app.AccountRole{}); res.Error != nil {
		return errors.Wrap(res.Error, "unable to remove stack role bindings")
	}

	if res := tx.Unscoped().
		Where("role_id IN ?", roleIDs).
		Delete(&app.Policy{}); res.Error != nil {
		return errors.Wrap(res.Error, "unable to delete stack role policies")
	}

	if res := tx.Unscoped().
		Where("id IN ?", roleIDs).
		Delete(&app.Role{}); res.Error != nil {
		return errors.Wrap(res.Error, "unable to delete stack roles")
	}

	return nil
}
