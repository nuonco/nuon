package migrations

import (
	"context"
	"fmt"
	"slices"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func legacyOrgPolicy(orgID string, name app.PolicyName, verb permissions.Permission, api permissions.API) app.Policy {
	return app.Policy{
		OrgID: generics.NewNullString(orgID),
		Name:  name,
		Permissions: pgtype.Hstore(map[string]*string{
			orgID:                        verb.ToStrPtr(),
			permissions.APIPermissionKey: generics.ToPtr(string(api)),
		}),
	}
}

func legacyOrgRoles(orgID string) []app.Role {
	return []app.Role{
		{
			OrgID:       generics.NewNullString(orgID),
			RoleType:    app.RoleTypeOrgAdmin,
			Version:     app.RoleVersionV2,
			Title:       "Admin",
			Description: "Full access to the organization and all of its resources.",
			Contexts:    []string{app.RoleContextTeam, app.RoleContextServiceAccount, app.RoleContextAPIToken, app.RoleContextTrustPolicy},
			Managed:     true,
			Policies: []app.Policy{
				legacyOrgPolicy(orgID, app.PolicyNameOrgAdmin, permissions.PermissionAll, permissions.APICtl),
			},
		},

		{
			OrgID:       generics.NewNullString(orgID),
			RoleType:    app.RoleTypeOrgSupport,
			Version:     app.RoleVersionV2,
			Title:       "Support",
			Description: "Deprecated full-access role; no longer offered for new assignments.",
			Managed:     true,
			Policies: []app.Policy{
				legacyOrgPolicy(orgID, app.PolicyNameOrgSupport, permissions.PermissionAll, permissions.APICtl),
			},
		},

		{
			OrgID:       generics.NewNullString(orgID),
			RoleType:    app.RoleTypeOrgReadOnly,
			Version:     app.RoleVersionV2,
			Title:       "Read-only",
			Description: "Read-only access to the organization and its resources.",
			Contexts:    []string{app.RoleContextTeam, app.RoleContextServiceAccount, app.RoleContextAPIToken, app.RoleContextTrustPolicy},
			Managed:     true,
			Policies: []app.Policy{
				legacyOrgPolicy(orgID, app.PolicyNameOrgReadOnly, permissions.PermissionRead, permissions.APICtl),
			},
		},

		{
			OrgID:       generics.NewNullString(orgID),
			RoleType:    app.RoleTypeInstaller,
			Version:     app.RoleVersionV2,
			Title:       "Installer",
			Description: "Deprecated full-access role; no longer offered for new assignments.",
			Managed:     true,
			Policies: []app.Policy{
				legacyOrgPolicy(orgID, app.PolicyNameInstaller, permissions.PermissionAll, permissions.APICtl),
			},
		},

		{
			OrgID:       generics.NewNullString(orgID),
			RoleType:    app.RoleTypeRunner,
			Version:     app.RoleVersionV2,
			Title:       "Runner",
			Description: "Permissions for runners executing deployments.",
			Managed:     true,
			Policies: []app.Policy{
				legacyOrgPolicy(orgID, "admin", permissions.PermissionAll, permissions.APIRunner),
			},
		},
	}
}

func reconcileLegacyRoleMetadata(ctx context.Context, db *gorm.DB, org app.Org) error {
	var roles []app.Role
	if err := db.WithContext(ctx).Where(app.Role{OrgID: generics.NewNullString(org.ID)}).
		Where(db.Where(app.Role{Version: app.RoleVersionV1}).Or(app.Role{Version: app.RoleVersionV2})).Find(&roles).Error; err != nil {
		return err
	}
	byType := make(map[app.RoleType]app.Role, len(roles))
	for _, role := range roles {
		byType[role.RoleType] = role
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, want := range legacyOrgRoles(org.ID) {
			existing, ok := byType[want.RoleType]
			if !ok {
				want.CreatedByID = org.CreatedByID
				for i := range want.Policies {
					want.Policies[i].CreatedByID = org.CreatedByID
				}
				if err := tx.Create(&want).Error; err != nil {
					return err
				}
				continue
			}
			if existing.Title == want.Title && existing.Description == want.Description &&
				existing.Managed == want.Managed && slices.Equal(existing.Contexts, want.Contexts) {
				continue
			}
			if err := tx.Model(&existing).Select("title", "description", "contexts", "managed").Updates(app.Role{
				Title: want.Title, Description: want.Description, Contexts: want.Contexts, Managed: want.Managed,
			}).Error; err != nil {
				return fmt.Errorf("unable to update %s role metadata for org %s: %w", want.RoleType, org.ID, err)
			}
		}
		return nil
	})
}
