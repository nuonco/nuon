package authz

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

type OrgRoleSnapshot struct {
	Role     app.Role
	Bindings []app.AccountRole
}

type OrgRoleClassification struct {
	Status      string
	Reason      string
	Permissions permissions.PermissionSet
	LegacyAPI   *permissions.API
	Policy      *permissions.Policy
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
		switch role.RoleType {
		case app.RoleTypeRunnerInstall, app.RoleTypeRunnerJob, app.RoleTypeStack, app.RoleTypeTelemetryCollector, app.RoleTypePhoneHome, app.RoleTypeLogStream:
		case app.RoleTypeOrgAdmin, app.RoleTypeOrgReadOnly, app.RoleTypeOrgSupport, app.RoleTypeInstaller, app.RoleTypeRunner:
		default:
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
	if err := scope.ValidateShape(); err != nil {
		return skip("invalid_org_scope")
	}
	policy := role.Policies[0]
	if policy.ID == "" || policy.DeletedAt != 0 || policy.RoleID != role.ID || policy.OrgID.Empty() || policy.OrgID.String != orgID {
		return skip("policy_owner_mismatch")
	}
	seenAccounts := make(map[string]bool, len(source.Bindings))
	for _, binding := range source.Bindings {
		if binding.AccountID == "" || binding.Account.ID != binding.AccountID || binding.Account.DeletedAt != 0 {
			continue
		}
		if role.Org == nil || role.Org.ID != orgID || role.Org.DeletedAt != 0 {
			return skip("missing_or_deleted_org")
		}
		if binding.DeletedAt != 0 || binding.RoleID != role.ID || binding.OrgID.Empty() || binding.OrgID.String != orgID {
			return skip("binding_owner_mismatch")
		}
		if seenAccounts[binding.AccountID] {
			return skip("duplicate_binding")
		}
		seenAccounts[binding.AccountID] = true
	}
	if hasDocument {
		return OrgRoleClassification{Status: "already_v3"}, nil
	}
	switch role.RoleType {
	case app.RoleTypeRunnerInstall, app.RoleTypeRunnerJob, app.RoleTypeStack, app.RoleTypeTelemetryCollector, app.RoleTypePhoneHome, app.RoleTypeLogStream:
		document, err := ScopedRolePolicy(&role)
		if err != nil {
			return skip("legacy_grant_mismatch")
		}
		grants, err := document.Compile(orgID, scope)
		if err != nil {
			return OrgRoleClassification{}, err
		}
		var api *permissions.API
		if value := document.API(); value != "" {
			api = &value
		}
		return OrgRoleClassification{Status: "eligible", Permissions: grants[0].Permissions, LegacyAPI: api, Policy: &document}, nil
	}
	verb := permissions.PermissionAll
	bundleName := permissions.RoleOrgAdmin
	if role.RoleType == app.RoleTypeOrgReadOnly {
		verb, bundleName = permissions.PermissionRead, permissions.RoleOrgReader
	}
	keys := 1
	var api *permissions.API
	if role.Version == app.RoleVersionV2 {
		keys++
		value := policy.Permissions[permissions.APIPermissionKey]
		expected := permissions.APICtl
		if role.RoleType == app.RoleTypeRunner {
			expected = permissions.APIRunner
		}
		if value == nil || *value != string(expected) {
			return skip("legacy_api_mismatch")
		}
		api = &expected
	}
	value := policy.Permissions[orgID]
	if len(policy.Permissions) != keys || value == nil || *value != string(verb) {
		return skip("legacy_grant_mismatch")
	}
	var retainedAPI permissions.API
	if api != nil {
		retainedAPI = *api
	}
	bundle, err := permissions.RolePolicy(bundleName, retainedAPI)
	if err != nil {
		return OrgRoleClassification{}, err
	}
	grants, err := bundle.Compile(orgID, scope)
	if err != nil {
		return OrgRoleClassification{}, err
	}
	return OrgRoleClassification{Status: "eligible", Permissions: grants[0].Permissions, LegacyAPI: api, Policy: &bundle}, nil
}
