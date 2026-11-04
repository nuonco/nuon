package authz

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

type OrgRoleSnapshot struct {
	Role     app.Role
	OrgRoles []app.Role
	Bindings []app.AccountRole
	Private  bool
}

type OrgRoleClassification struct {
	Status      string
	Reason      string
	Permissions permissions.PermissionSet
	LegacyAPI   *permissions.API
}

func ClassifyOrgRole(source OrgRoleSnapshot) (OrgRoleClassification, error) {
	skip := func(reason string) (OrgRoleClassification, error) {
		return OrgRoleClassification{Status: "skipped", Reason: reason}, nil
	}
	role := source.Role
	if role.ID == "" || role.DeletedAt != 0 {
		return skip("missing_or_deleted_role")
	}
	hasDocument := role.HasPolicyDocument()
	for _, policy := range role.Policies {
		if policy.Document != nil {
			if _, err := policy.PermissionProjection(); err != nil {
				return OrgRoleClassification{}, fmt.Errorf("role %s: %w", role.ID, err)
			}
		}
	}
	if !hasDocument {
		if source.Private {
			return skip("private_role")
		}
		if role.RoleType != app.RoleTypeOrgAdmin && role.RoleType != app.RoleTypeOrgReadOnly {
			return skip("unsupported_role_type")
		}
		if role.Version != app.RoleVersionV1 && role.Version != app.RoleVersionV2 {
			return skip("unsupported_legacy_version")
		}
	}
	if len(role.Policies) != 1 {
		if hasDocument {
			return OrgRoleClassification{}, fmt.Errorf("role %s requires one live document policy", role.ID)
		}
		return skip("policy_count")
	}
	orgID := role.OrgID.ValueString()
	scope := permissions.Scope{Kind: permissions.KindOrg, OrgID: orgID, ID: orgID}
	if err := scope.ValidateShape(); err != nil || role.Org == nil || role.Org.ID != orgID || role.Org.DeletedAt != 0 {
		return skip("missing_or_deleted_org")
	}
	policy := role.Policies[0]
	if policy.ID == "" || policy.DeletedAt != 0 || policy.RoleID != role.ID || policy.OrgID.Empty() || policy.OrgID.String != orgID {
		return skip("policy_owner_mismatch")
	}
	seenAccounts := make(map[string]bool, len(source.Bindings))
	for _, binding := range source.Bindings {
		if binding.DeletedAt != 0 || binding.RoleID != role.ID || binding.OrgID.Empty() || binding.OrgID.String != orgID {
			return skip("binding_owner_mismatch")
		}
		if binding.AccountID == "" || binding.Account.ID != binding.AccountID || binding.Account.DeletedAt != 0 {
			return skip("missing_or_deleted_account")
		}
		if seenAccounts[binding.AccountID] {
			return skip("duplicate_binding")
		}
		seenAccounts[binding.AccountID] = true
	}
	if hasDocument {
		return OrgRoleClassification{Status: "already_v3"}, nil
	}
	standardRoles := 0
	foundRole := false
	for _, other := range source.OrgRoles {
		if other.DeletedAt == 0 && other.OrgID.ValueString() == orgID && other.RoleType == role.RoleType {
			standardRoles++
			foundRole = foundRole || other.ID == role.ID
		}
	}
	if standardRoles != 1 || !foundRole {
		return skip("ambiguous_standard_role")
	}
	verb := permissions.PermissionRead
	if role.RoleType == app.RoleTypeOrgAdmin {
		verb = permissions.PermissionAll
	}
	keys := 1
	var api *permissions.API
	if role.Version == app.RoleVersionV2 {
		keys++
		value := policy.Permissions[permissions.APIPermissionKey]
		if value == nil || *value != string(permissions.APICtl) {
			return skip("legacy_api_mismatch")
		}
		ctl := permissions.APICtl
		api = &ctl
	}
	value := policy.Permissions[orgID]
	if len(policy.Permissions) != keys || value == nil || *value != string(verb) {
		return skip("legacy_grant_mismatch")
	}
	for _, binding := range source.Bindings {
		for _, other := range binding.Account.Roles {
			if other.DeletedAt != 0 {
				continue
			}
			for _, otherPolicy := range other.Policies {
				if otherPolicy.DeletedAt != 0 {
					continue
				}
				projection, err := otherPolicy.PermissionProjection()
				if err != nil {
					return OrgRoleClassification{}, fmt.Errorf("account %s role %s: %w", binding.AccountID, other.ID, err)
				}
				if verb != permissions.PermissionRead {
					continue
				}
				value, exists := projection[orgID]
				if !exists {
					continue
				}
				if value == nil {
					return skip("invalid_holder_grant")
				}
				switch permissions.Permission(*value) {
				case permissions.PermissionAll, permissions.PermissionRead:
				case permissions.PermissionCreate, permissions.PermissionUpdate, permissions.PermissionDelete:
					return skip("order_sensitive_grants")
				default:
					return skip("invalid_holder_grant")
				}
			}
		}
	}
	bundle, err := permissions.RolePolicy(permissions.RoleName(role.RoleType), permissions.APICtl)
	if err != nil {
		return OrgRoleClassification{}, err
	}
	grants, err := bundle.Compile(orgID, scope)
	if err != nil {
		return OrgRoleClassification{}, err
	}
	return OrgRoleClassification{Status: "eligible", Permissions: grants[0].Permissions, LegacyAPI: api}, nil
}
