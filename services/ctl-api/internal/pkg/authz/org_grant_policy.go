package authz

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func orgGrantPolicy(orgID string, roleType app.RoleType, api permissions.API) ([]byte, error) {
	refs := []permissions.PermissionID{permissions.OrgRead, permissions.SlackConnect}
	switch roleType {
	case app.RoleTypeOrgReadOnly:
	case app.RoleTypeOrgAdmin:
		refs = []permissions.PermissionID{permissions.OrgRead, permissions.OrgWrite, permissions.ServiceAccountsRead, permissions.TokensRead, permissions.TrustPoliciesRead}
	default:
		return nil, fmt.Errorf("role %s has no customer grant recipe", roleType)
	}
	catalog, err := permissions.DefaultCatalog()
	if err != nil {
		return nil, err
	}
	document, err := catalog.NormalizePolicy(permissions.Policy{
		Format: permissions.PolicyFormatV4, API: api,
		Grants: []permissions.Grant{{Scope: permissions.GrantScope{Kind: permissions.KindOrg, IDs: []string{orgID}}, Permissions: refs}},
	})
	if err != nil {
		return nil, err
	}
	if _, err := catalog.Compile(document, orgID); err != nil {
		return nil, err
	}
	return json.Marshal(document)
}

func useOrgGrantPolicies(roles []app.Role) error {
	for i := range roles {
		role := &roles[i]
		if role.RoleType != app.RoleTypeOrgAdmin && role.RoleType != app.RoleTypeOrgReadOnly {
			continue
		}
		document, err := orgGrantPolicy(role.OrgID.ValueString(), role.RoleType, "")
		if err != nil {
			return err
		}
		role.Policies[0].Document = document
		role.Policies[0].Permissions = nil
	}
	return nil
}

func isStandardOrgGrant(role app.Role) bool {
	if !role.Managed || len(role.Policies) != 1 || !role.Policies[0].HasGrantDocument() {
		return false
	}
	catalog, err := permissions.DefaultCatalog()
	if err != nil {
		return false
	}
	document, err := role.Policies[0].DecodeGrants(catalog)
	if err != nil || (document.API != "" && document.API != permissions.APICtl) {
		return false
	}
	expected, err := orgGrantPolicy(role.OrgID.ValueString(), role.RoleType, document.API)
	if err != nil {
		return false
	}
	actual, err := json.Marshal(document)
	return err == nil && bytes.Equal(expected, actual)
}
