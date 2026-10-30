package app

import (
	"fmt"
	"maps"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func (r Role) ValidateLegacyCompatibility() error {
	if r.Version != RoleVersionV3 || (r.LegacyVersion != RoleVersionV1 && r.LegacyVersion != RoleVersionV2) ||
		!r.Managed || r.ID == "" || len(r.Policies) != 1 {
		return fmt.Errorf("role is not a legacy-compatible v3 catalog role")
	}
	verb := permissions.PermissionRead
	switch r.RoleType {
	case RoleTypeOrgAdmin:
		verb = permissions.PermissionAll
	case RoleTypeOrgReadOnly:
	default:
		return fmt.Errorf("role type %q has no legacy compatibility contract", r.RoleType)
	}
	orgID := r.OrgID.ValueString()
	scope := permissions.Scope{Kind: permissions.KindOrg, OrgID: orgID, ID: orgID}
	if err := scope.ValidateShape(); err != nil {
		return err
	}
	stored := r.Policies[0]
	if stored.RoleID != r.ID || stored.OrgID.ValueString() != orgID {
		return fmt.Errorf("compatible policy must belong to its role and tenant")
	}
	grant := stored.Permissions[orgID]
	if grant == nil || *grant != string(verb) {
		return fmt.Errorf("compatible role has an unexpected legacy grant")
	}
	wantKeys := 1
	if r.LegacyVersion == RoleVersionV2 {
		api := stored.Permissions[permissions.APIPermissionKey]
		if api == nil || *api != string(permissions.APICtl) {
			return fmt.Errorf("compatible v2 role requires ctl-api eligibility")
		}
		wantKeys++
	}
	if len(stored.Permissions) != wantKeys {
		return fmt.Errorf("compatible role contains additional legacy authority")
	}
	policy, err := permissions.DecodePolicy(r.Version, stored.Document)
	if err != nil {
		return err
	}
	if policy.Kind() != permissions.PolicyKindBundle || policy.API() != permissions.APICtl {
		return fmt.Errorf("compatible role requires a public API bundle")
	}
	grants, err := policy.Compile(orgID, scope)
	if err != nil {
		return err
	}
	definition, err := permissions.CurrentCustomerRole(permissions.CustomerRoleName(r.RoleType))
	if err != nil {
		return err
	}
	if r.CatalogRevision != definition.Revision || len(grants) != 1 || !maps.Equal(grants[0].Permissions, definition.Permissions) {
		return fmt.Errorf("compatible role differs from its versioned catalog bundle")
	}
	return nil
}

func (b AccountRole) ValidateLegacyCompatibility() error {
	if err := b.Role.ValidateLegacyCompatibility(); err != nil {
		return err
	}
	orgID := b.Role.OrgID.ValueString()
	if b.RoleID != b.Role.ID || b.OrgID.ValueString() != orgID || b.ScopeKind != permissions.KindOrg || b.ScopeID != orgID {
		return fmt.Errorf("legacy-compatible role requires an explicit binding to its own org")
	}
	return nil
}
