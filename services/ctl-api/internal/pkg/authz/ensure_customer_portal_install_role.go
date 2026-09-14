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

// customerPortalInstallRole grants PermissionAll on the install because the HSTORE
// permission model stores one verb per object key and cannot express Read+Update+Create
// without Delete. Route-level require.Route middleware enforces the actual verbs
// (Read, Update, Create) — no Delete route is registered on the portal group. Adding
// one would expand portal SA access and must be reviewed here.
func customerPortalInstallRole(orgID, installID string) *app.Role {
	return &app.Role{
		OrgID:       generics.NewNullString(orgID),
		RoleType:    app.RoleTypeCustomerPortal,
		Title:       "Customer portal",
		Description: "Scoped access to customer portal endpoints for a single install.",
		Policies: []app.Policy{{
			OrgID: generics.NewNullString(orgID),
			Name:  app.PolicyNameCustomerPortal,
			Permissions: pgtype.Hstore(map[string]*string{
				permissions.Object(orgID, permissions.KindInstall, installID): permissions.PermissionAll.ToStrPtr(),
			}),
		}},
	}
}

func (h *Client) EnsureCustomerPortalInstallRole(ctx context.Context, orgID, installID, accountID string) error {
	ctx = cctx.SetAccountIDContext(ctx, accountID)

	var existing []app.Role
	if err := h.db.WithContext(ctx).
		Joins("JOIN account_roles ON account_roles.role_id = roles.id AND account_roles.deleted_at = 0").
		Preload("Policies").
		Where("account_roles.account_id = ?", accountID).
		Where("roles.org_id = ? AND roles.role_type = ?", orgID, app.RoleTypeCustomerPortal).
		Find(&existing).Error; err != nil {
		return errors.Wrap(err, "unable to look up customer portal role bindings")
	}
	want := permissions.Object(orgID, permissions.KindInstall, installID)
	wantVerb := string(permissions.PermissionAll)
	for _, role := range existing {
		if len(role.Policies) != 1 {
			continue
		}
		verb, ok := role.Policies[0].Permissions[want]
		if !ok {
			continue
		}
		if verb == nil || *verb != wantVerb {
			if err := h.db.WithContext(ctx).
				Model(&app.Policy{}).
				Where(app.Policy{ID: role.Policies[0].ID}).
				Update("permissions", pgtype.Hstore(map[string]*string{
					want: permissions.PermissionAll.ToStrPtr(),
				})).Error; err != nil {
				return errors.Wrap(err, "unable to converge customer portal role policy")
			}
		}
		return nil
	}

	role := customerPortalInstallRole(orgID, installID)
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(role).Error; err != nil {
			return errors.Wrap(err, "unable to create customer portal role")
		}
		binding := app.AccountRole{OrgID: generics.NewNullString(orgID), RoleID: role.ID, AccountID: accountID}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&binding).Error; err != nil {
			return errors.Wrap(err, "unable to bind customer portal role")
		}
		return nil
	})
}

func DeleteCustomerPortalInstallRoles(tx *gorm.DB, accountID string) error {
	var roleIDs []string
	if res := tx.Unscoped().
		Model(&app.Role{}).
		Joins("JOIN account_roles ON account_roles.role_id = roles.id").
		Where("account_roles.account_id = ?", accountID).
		Where("roles.role_type = ?", app.RoleTypeCustomerPortal).
		Distinct().
		Pluck("roles.id", &roleIDs); res.Error != nil {
		return errors.Wrap(res.Error, "unable to look up customer portal roles for account")
	}

	if len(roleIDs) == 0 {
		return nil
	}

	if res := tx.Unscoped().
		Where("role_id IN ?", roleIDs).
		Delete(&app.AccountRole{}); res.Error != nil {
		return errors.Wrap(res.Error, "unable to remove customer portal role bindings")
	}

	if res := tx.Unscoped().
		Where("role_id IN ?", roleIDs).
		Delete(&app.Policy{}); res.Error != nil {
		return errors.Wrap(res.Error, "unable to delete customer portal role policies")
	}

	if res := tx.Unscoped().
		Where("id IN ?", roleIDs).
		Delete(&app.Role{}); res.Error != nil {
		return errors.Wrap(res.Error, "unable to delete customer portal roles")
	}

	return nil
}
