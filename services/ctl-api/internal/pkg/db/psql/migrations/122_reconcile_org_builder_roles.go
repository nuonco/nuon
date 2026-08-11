package migrations

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func (m *Migrations) Migration122ReconcileOrgBuilderRoles(ctx context.Context, db *gorm.DB) error {
	var orgs []app.Org
	if err := db.WithContext(ctx).Find(&orgs).Error; err != nil {
		return fmt.Errorf("unable to fetch orgs: %w", err)
	}

	for _, org := range orgs {
		if err := reconcileOrgBuilderRole(ctx, db, org); err != nil {
			return fmt.Errorf("unable to reconcile org_builder role for org %s: %w", org.ID, err)
		}
	}

	return nil
}

func reconcileOrgBuilderRole(ctx context.Context, db *gorm.DB, org app.Org) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var roles []app.Role
		if err := tx.
			Where(app.Role{OrgID: generics.NewNullString(org.ID), RoleType: app.RoleTypeOrgBuilder}).
			Find(&roles).Error; err != nil {
			return err
		}
		if len(roles) == 0 {
			role := newOrgBuilderRole(org)
			return tx.Create(&role).Error
		}

		for _, role := range roles {
			var policy app.Policy
			err := tx.Unscoped().
				Where(app.Policy{RoleID: role.ID}).
				First(&policy).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				policy = newOrgBuilderPolicy(org, role.ID)
				if err := tx.Create(&policy).Error; err != nil {
					return err
				}
				continue
			}
			if err != nil {
				return err
			}

			if err := tx.Unscoped().Model(&policy).Updates(map[string]any{
				"deleted_at":  soft_delete.DeletedAt(0),
				"name":        app.PolicyNameOrgBuilder,
				"org_id":      generics.NewNullString(org.ID),
				"permissions": orgBuilderPermissions(org.ID),
			}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func newOrgBuilderRole(org app.Org) app.Role {
	return app.Role{
		OrgID:       generics.NewNullString(org.ID),
		CreatedByID: org.CreatedByID,
		RoleType:    app.RoleTypeOrgBuilder,
		Policies:    []app.Policy{newOrgBuilderPolicy(org, "")},
	}
}

func newOrgBuilderPolicy(org app.Org, roleID string) app.Policy {
	return app.Policy{
		RoleID:      roleID,
		OrgID:       generics.NewNullString(org.ID),
		CreatedByID: org.CreatedByID,
		Name:        app.PolicyNameOrgBuilder,
		Permissions: orgBuilderPermissions(org.ID),
	}
}

func orgBuilderPermissions(orgID string) pgtype.Hstore {
	return pgtype.Hstore(map[string]*string{
		orgID:                                  permissions.PermissionRead.ToStrPtr(),
		orgBuilderComponentBuildsObject(orgID): permissions.PermissionCreate.ToStrPtr(),
	})
}
