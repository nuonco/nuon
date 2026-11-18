package authz

import (
	"encoding/json"
	"fmt"
	"reflect"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func workloadGrantPolicy(role *app.Role) (permissions.Policy, error) {
	orgID := role.OrgID.ValueString()
	if role.RoleType == app.RoleTypeRunner {
		if !role.Managed || len(role.Contexts) != 0 || len(role.Policies) != 1 {
			return permissions.Policy{}, fmt.Errorf("runner requires its managed held-only role")
		}
		projection, err := role.Policies[0].PermissionProjection()
		if err != nil {
			return permissions.Policy{}, err
		}
		if len(projection) != 2 || projection[orgID] == nil || *projection[orgID] != string(permissions.PermissionAll) ||
			projection[permissions.APIPermissionKey] == nil || *projection[permissions.APIPermissionKey] != string(permissions.APIRunner) {
			return permissions.Policy{}, fmt.Errorf("runner role differs from its runner-API baseline")
		}
		return workloadGrantRecipe(role.RoleType, permissions.Scope{Kind: permissions.KindOrg, OrgID: orgID, ID: orgID}, permissions.APIRunner)
	}
	legacy, err := ScopedRolePolicy(role)
	if err != nil {
		return permissions.Policy{}, err
	}
	rules, err := legacy.Compile(orgID, permissions.Scope{Kind: permissions.KindOrg, OrgID: orgID, ID: orgID})
	if err != nil {
		return permissions.Policy{}, err
	}
	if len(rules) != 1 {
		return permissions.Policy{}, fmt.Errorf("private role requires one resource rule")
	}
	return workloadGrantRecipe(role.RoleType, rules[0].Scope, legacy.API())
}

func workloadGrantRecipe(roleType app.RoleType, scope permissions.Scope, api permissions.API) (permissions.Policy, error) {
	var refs []permissions.PermissionID
	var kind permissions.ResourceKind
	switch roleType {
	case app.RoleTypeRunner:
		kind = permissions.KindOrg
		refs = append(runnerInstallPermissions(),
			permissions.OrgInspect, permissions.RunnerInspect, permissions.RunnerHealthReport,
			permissions.RunnerComponentHealthRead, permissions.RunnerComponentHealthReport,
			permissions.RunnerQueueRead, permissions.RunnerProcessReport, permissions.RunnerTelemetryWrite,
			permissions.RunnerJobInspect, permissions.RunnerJobReport, permissions.RunnerJobClaim,
			permissions.StackConfigRead, permissions.StackReport,
			permissions.TelemetrySettingsRead, permissions.TelemetryTokenCreate,
		)
	case app.RoleTypeRunnerInstall:
		kind = permissions.KindInstall
		refs = runnerInstallPermissions()
	case app.RoleTypeRunnerJob:
		kind = permissions.KindRunnerJob
		refs = []permissions.PermissionID{permissions.RunnerJobInspect, permissions.RunnerJobReport}
	case app.RoleTypeStack:
		kind = permissions.KindInstall
		refs = []permissions.PermissionID{permissions.StackConfigRead, permissions.StackReport}
	case app.RoleTypeTelemetryCollector:
		kind = permissions.KindInstall
		refs = []permissions.PermissionID{permissions.TelemetrySettingsRead, permissions.TelemetryTokenCreate}
	case app.RoleTypePhoneHome:
		kind = permissions.KindPhoneHome
		refs = []permissions.PermissionID{permissions.StackVersionReport}
	case app.RoleTypeLogStream:
		kind = permissions.KindLogStream
		refs = []permissions.PermissionID{permissions.LogStreamAppend}
	default:
		return permissions.Policy{}, fmt.Errorf("unsupported workload grant role %s", roleType)
	}
	if scope.Kind != kind || scope.ID == "*" {
		return permissions.Policy{}, fmt.Errorf("workload role requires one explicit %s", kind)
	}
	catalog, err := permissions.DefaultCatalog()
	if err != nil {
		return permissions.Policy{}, err
	}
	policy, err := catalog.NormalizePolicy(permissions.Policy{
		Format: permissions.PolicyFormatV4, API: api,
		Grants: []permissions.Grant{{Scope: permissions.GrantScope{Kind: kind, IDs: []string{scope.ID}}, Permissions: refs}},
	})
	if err != nil {
		return permissions.Policy{}, err
	}
	_, err = catalog.Compile(policy, scope.OrgID)
	return policy, err
}

func runnerInstallPermissions() []permissions.PermissionID {
	return []permissions.PermissionID{
		permissions.ExecutionConfigRead, permissions.ActionRunReport, permissions.InstallActionInspect,
		permissions.RunnerTelemetryTokenCreate, permissions.RunnerSandboxConfigRead,
		permissions.TerraformRead, permissions.TerraformCreate, permissions.TerraformDelete,
		permissions.TerraformLock, permissions.TerraformStateWrite, permissions.TerraformStateDelete,
		permissions.HelmReleaseRead, permissions.HelmReleaseWrite, permissions.HelmReleaseDelete,
		permissions.LogStreamAppend, permissions.InstallHistoryRead,
	}
}

func decodeWorkloadGrant(role *app.Role) (permissions.Policy, error) {
	if len(role.Policies) != 1 {
		return permissions.Policy{}, fmt.Errorf("workload role requires one policy")
	}
	catalog, err := permissions.DefaultCatalog()
	if err != nil {
		return permissions.Policy{}, err
	}
	policy, err := role.Policies[0].DecodeGrants(catalog)
	if err != nil {
		return permissions.Policy{}, err
	}
	if len(policy.Grants) != 1 || len(policy.Grants[0].Scope.IDs) != 1 {
		return permissions.Policy{}, fmt.Errorf("workload role requires one explicit resource grant")
	}
	grant := policy.Grants[0]
	expected, err := workloadGrantRecipe(role.RoleType, permissions.Scope{
		Kind: grant.Scope.Kind, OrgID: role.OrgID.ValueString(), ID: grant.Scope.IDs[0],
	}, policy.API)
	if err != nil {
		return permissions.Policy{}, err
	}
	if !reflect.DeepEqual(policy, expected) {
		return permissions.Policy{}, fmt.Errorf("workload role differs from its managed grant recipe")
	}
	return policy, nil
}

func reconcilePrivateGrant(tx *gorm.DB, role, want *app.Role) error {
	current, err := decodeWorkloadGrant(role)
	if err != nil {
		return err
	}
	desired, err := workloadGrantPolicy(want)
	if err != nil {
		return err
	}
	desired.API = current.API
	if reflect.DeepEqual(current, desired) {
		return nil
	}
	document, err := json.Marshal(desired)
	if err != nil {
		return err
	}
	policy := &role.Policies[0]
	if err := tx.Model(&app.Policy{}).Where(app.Policy{ID: policy.ID}).
		Updates(map[string]any{"document": document, "permissions": nil}).Error; err != nil {
		return err
	}
	policy.Document, policy.Permissions = document, nil
	return nil
}

func (h *Client) privateGrantIssuance(tx *gorm.DB, accountID, orgID string) (bool, error) {
	var roles []app.Role
	if err := tx.Joins("JOIN account_roles ON account_roles.role_id = roles.id AND account_roles.deleted_at = 0").
		Preload("Policies").Where("account_roles.account_id = ?", accountID).
		Where(app.Role{OrgID: generics.NewNullString(orgID)}).Find(&roles).Error; err != nil {
		return false, err
	}
	if len(roles) == 0 {
		return h.cfg != nil && h.cfg.EnableOrgGrantPolicies, nil
	}
	grants := len(roles[0].Policies) == 1 && roles[0].Policies[0].HasGrantDocument()
	for _, role := range roles {
		if (len(role.Policies) == 1 && role.Policies[0].HasGrantDocument()) != grants {
			return false, fmt.Errorf("account mixes grant and legacy roles")
		}
	}
	return grants, nil
}
