package authz

import (
	"fmt"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

type scopedGrant struct {
	objectKind, scopeKind permissions.ResourceKind
	permission            permissions.PermissionID
	verb                  permissions.Permission
}

var (
	installGrant   = scopedGrant{permissions.KindInstall, permissions.KindInstall, permissions.InstallAll, permissions.PermissionAll}
	runnerJobGrant = scopedGrant{permissions.KindRunnerJob, permissions.KindRunnerJob, permissions.RunnerJobAll, permissions.PermissionAll}
)

// scopedRoleGrants lists the grants a scoped role type may hold. The first is
// required; the rest are optional, such as a runner job's install.
func scopedRoleGrants(roleType app.RoleType) ([]scopedGrant, error) {
	switch roleType {
	case app.RoleTypeRunnerInstall:
		return []scopedGrant{installGrant}, nil
	case app.RoleTypeStack:
		return []scopedGrant{{permissions.KindStack, permissions.KindInstall, permissions.StackAll, permissions.PermissionAll}}, nil
	case app.RoleTypeTelemetryCollector:
		return []scopedGrant{{permissions.KindTelemetry, permissions.KindInstall, permissions.TelemetryAll, permissions.PermissionAll}}, nil
	case app.RoleTypeRunnerJob:
		return []scopedGrant{runnerJobGrant, installGrant}, nil
	case app.RoleTypePhoneHome:
		return []scopedGrant{{permissions.KindPhoneHome, permissions.KindPhoneHome, permissions.PhoneHomeCreate, permissions.PermissionCreate}}, nil
	case app.RoleTypeLogStream:
		return []scopedGrant{{permissions.KindLogStream, permissions.KindLogStream, permissions.LogStreamCreate, permissions.PermissionCreate}}, nil
	default:
		return nil, fmt.Errorf("unsupported scoped role type %q", roleType)
	}
}

func ScopedRolePolicy(role *app.Role) (permissions.Policy, error) {
	if len(role.Policies) != 1 {
		return permissions.Policy{}, fmt.Errorf("scoped role requires one policy")
	}
	projection, err := role.Policies[0].PermissionProjection()
	if err != nil {
		return permissions.Policy{}, err
	}
	api, enforced, err := APIFromRole(role)
	if err != nil {
		return permissions.Policy{}, err
	}
	orgID := role.OrgID.String
	if role.RoleType == app.RoleTypeRunnerIdentity {
		keys := 1
		if enforced {
			keys++
		}
		if grant := projection[orgID]; grant == nil || *grant != string(permissions.PermissionAll) || len(projection) != keys {
			return permissions.Policy{}, fmt.Errorf("runner identity requires full access to its owning org")
		}
		return permissions.RolePolicy(permissions.RoleOrgAdmin, api)
	}
	allowed, err := scopedRoleGrants(role.RoleType)
	if err != nil {
		return permissions.Policy{}, err
	}
	matched := 0
	grants := make([]permissions.ResourceGrant, 0, len(allowed))
	for i, grant := range allowed {
		prefix := permissions.Object(orgID, grant.objectKind, "")
		var ids []string
		for object, value := range projection {
			if object == permissions.APIPermissionKey || !strings.HasPrefix(object, prefix) {
				continue
			}
			if value == nil || *value != string(grant.verb) {
				return permissions.Policy{}, fmt.Errorf("scoped role has an unexpected resource grant")
			}
			ids = append(ids, strings.TrimPrefix(object, prefix))
		}
		if len(ids) > 1 {
			return permissions.Policy{}, fmt.Errorf("scoped role has an unexpected policy shape")
		}
		if i == 0 && len(ids) == 0 {
			return permissions.Policy{}, fmt.Errorf("scoped role requires a resource grant")
		}
		for _, id := range ids {
			matched++
			grants = append(grants, permissions.ResourceGrant{
				Scope: permissions.Scope{Kind: grant.scopeKind, OrgID: orgID, ID: id}, Permission: grant.permission,
			})
		}
	}
	keys := matched
	if enforced {
		keys++
	}
	if len(projection) != keys {
		return permissions.Policy{}, fmt.Errorf("scoped role has an unexpected policy shape")
	}
	return permissions.ResourcePolicy(api, grants...)
}
