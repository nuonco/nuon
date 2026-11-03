package authz

import (
	"context"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func (c *Client) CreateOrgRoles(ctx context.Context, orgID string) error {
	return ReconcileOrgRoles(ctx, c.db, app.Org{ID: orgID})
}

func standardOrgRoles(orgID string) ([]app.Role, error) {
	definitions := []struct {
		name               permissions.CustomerRoleName
		title, description string
	}{
		{permissions.RoleOrgReader, "Org Reader", "Read the organization and its resources."},
		{permissions.RoleOrgEditor, "Org Editor", "Read and operate apps and installs without deleting them or administering access."},
		{permissions.RoleOrgAdmin, "Org Admin", "Administer organization access and operate or delete its apps and installs."},
		{permissions.RoleAppReader, "App Reader", "Read the assigned app and its installs."},
		{permissions.RoleAppEditor, "App Editor", "Operate the assigned app and its installs without deleting them or administering access."},
		{permissions.RoleAppAdmin, "App Admin", "Operate or delete the assigned app and its installs without administering access."},
		{permissions.RoleInstallReader, "Install Reader", "Read the assigned install."},
		{permissions.RoleInstallEditor, "Install Editor", "Operate the assigned install without deleting it or administering access."},
		{permissions.RoleInstallAdmin, "Install Admin", "Operate, delete, forget or fully deprovision the assigned install without administering access."},
	}
	roles := make([]app.Role, 0, len(definitions))
	for _, definition := range definitions {
		bundle, err := permissions.CurrentCustomerRole(definition.name)
		if err != nil {
			return nil, err
		}
		document, err := bundle.MarshalPolicy()
		if err != nil {
			return nil, err
		}
		roles = append(roles, app.Role{
			OrgID: generics.NewNullString(orgID), RoleType: app.RoleType(bundle.Name),
			Version: app.RoleVersionV3, CatalogRevision: bundle.Revision,
			Title: definition.title, Description: definition.description, Managed: true,
			Contexts: []string{app.RoleContextTeam, app.RoleContextServiceAccount, app.RoleContextAPIToken, app.RoleContextTrustPolicy},
			Policies: []app.Policy{{OrgID: generics.NewNullString(orgID), Name: app.PolicyName(bundle.Name), Document: document}},
		})
	}
	return roles, nil
}
