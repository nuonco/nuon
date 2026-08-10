package authz

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func (c *Client) CreateOrgRoles(ctx context.Context, orgID string) error {
	roles := standardOrgRoles(orgID)

	res := c.db.
		WithContext(ctx).
		Create(roles)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

// standardOrgRoles defines the managed roles every org gets. It is the single
// source of truth for their permissions and metadata: new orgs get these rows
// at creation, and ReconcileOrgRoles keeps existing orgs' metadata in sync.
func standardOrgRoles(orgID string) []app.Role {
	return []app.Role{
		{
			OrgID:       generics.NewNullString(orgID),
			RoleType:    app.RoleTypeOrgAdmin,
			Title:       "Admin",
			Description: "Full access to the organization and all of its resources.",
			Contexts:    []string{app.RoleContextTeam, app.RoleContextServiceAccount, app.RoleContextAPIToken, app.RoleContextTrustPolicy},
			Managed:     true,
			Policies: []app.Policy{
				{
					OrgID: generics.NewNullString(orgID),
					Name:  app.PolicyNameOrgAdmin,
					Permissions: pgtype.Hstore(map[string]*string{
						orgID: permissions.PermissionAll.ToStrPtr(),
					}),
				},
			},
		},

		{
			OrgID:       generics.NewNullString(orgID),
			RoleType:    app.RoleTypeOrgSupport,
			Title:       "Support",
			Description: "Deprecated full-access role; no longer offered for new assignments.",
			Managed:     true,
			Policies: []app.Policy{
				{
					OrgID: generics.NewNullString(orgID),
					Name:  app.PolicyNameOrgSupport,
					Permissions: pgtype.Hstore(map[string]*string{
						orgID: permissions.PermissionAll.ToStrPtr(),
					}),
				},
			},
		},

		{
			OrgID:       generics.NewNullString(orgID),
			RoleType:    app.RoleTypeOrgReadOnly,
			Title:       "Read-only",
			Description: "Read-only access to the organization and its resources.",
			Contexts:    []string{app.RoleContextTeam, app.RoleContextServiceAccount, app.RoleContextAPIToken, app.RoleContextTrustPolicy},
			Managed:     true,
			Policies: []app.Policy{
				{
					OrgID: generics.NewNullString(orgID),
					Name:  app.PolicyNameOrgReadOnly,
					Permissions: pgtype.Hstore(map[string]*string{
						orgID: permissions.PermissionRead.ToStrPtr(),
					}),
				},
			},
		},

		{
			OrgID:       generics.NewNullString(orgID),
			RoleType:    app.RoleTypeInstaller,
			Title:       "Installer",
			Description: "Deprecated full-access role; no longer offered for new assignments.",
			Managed:     true,
			Policies: []app.Policy{
				{
					OrgID: generics.NewNullString(orgID),
					Name:  app.PolicyNameInstaller,
					Permissions: pgtype.Hstore(map[string]*string{
						orgID: permissions.PermissionAll.ToStrPtr(),
					}),
				},
			},
		},

		{
			OrgID:       generics.NewNullString(orgID),
			RoleType:    app.RoleTypeRunner,
			Title:       "Runner",
			Description: "Permissions for runners executing deployments.",
			Managed:     true,
			Policies: []app.Policy{
				{
					OrgID: generics.NewNullString(orgID),
					Name:  "admin",
					Permissions: pgtype.Hstore(map[string]*string{
						orgID: permissions.PermissionAll.ToStrPtr(),
					}),
				},
			},
		},
	}
}
