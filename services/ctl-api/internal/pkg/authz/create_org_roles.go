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
func orgPolicy(orgID string, name app.PolicyName, verb permissions.Permission, api permissions.API) app.Policy {
	return app.Policy{
		OrgID: generics.NewNullString(orgID),
		Name:  name,
		Permissions: pgtype.Hstore(map[string]*string{
			orgID:                        verb.ToStrPtr(),
			permissions.APIPermissionKey: generics.ToPtr(string(api)),
		}),
	}
}

func standardOrgRoles(orgID string) []app.Role {
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
				orgPolicy(orgID, app.PolicyNameOrgAdmin, permissions.PermissionAll, permissions.APICtl),
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
				orgPolicy(orgID, app.PolicyNameOrgSupport, permissions.PermissionAll, permissions.APICtl),
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
				orgPolicy(orgID, app.PolicyNameOrgReadOnly, permissions.PermissionRead, permissions.APICtl),
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
				orgPolicy(orgID, app.PolicyNameInstaller, permissions.PermissionAll, permissions.APICtl),
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
				orgPolicy(orgID, "admin", permissions.PermissionAll, permissions.APIRunner),
			},
		},
	}
}
