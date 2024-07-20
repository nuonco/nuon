package authz

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz/permissions"
)

func (c *Client) CreateOrgRoles(ctx context.Context, orgID string) error {
	roles := []app.Role{
		// create admin role
		{
			OrgID:    generics.NewNullString(orgID),
			RoleType: app.RoleTypeOrgAdmin,
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

		// installer role
		{
			OrgID:    generics.NewNullString(orgID),
			RoleType: app.RoleTypeInstaller,
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

		// runner role
		{
			OrgID:    generics.NewNullString(orgID),
			RoleType: app.RoleTypeRunner,
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

	res := c.db.
		WithContext(ctx).
		Create(roles)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
