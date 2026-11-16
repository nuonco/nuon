package authz

import (
	"fmt"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func ScopedRolePolicy(role *app.Role) (permissions.LegacyPolicy, error) {
	if len(role.Policies) != 1 {
		return permissions.LegacyPolicy{}, fmt.Errorf("scoped role requires one policy")
	}
	var objectKind, scopeKind permissions.ResourceKind
	var permission permissions.PermissionID
	verb := permissions.PermissionAll
	switch role.RoleType {
	case app.RoleTypeRunnerInstall:
		objectKind, scopeKind, permission = permissions.KindInstall, permissions.KindInstall, permissions.InstallAll
	case app.RoleTypeStack:
		objectKind, scopeKind, permission = permissions.KindStack, permissions.KindInstall, permissions.StackAll
	case app.RoleTypeTelemetryCollector:
		objectKind, scopeKind, permission = permissions.KindTelemetry, permissions.KindInstall, permissions.TelemetryAll
	case app.RoleTypeRunnerJob:
		objectKind, scopeKind, permission = permissions.KindRunnerJob, permissions.KindRunnerJob, permissions.RunnerJobAll
	case app.RoleTypePhoneHome:
		objectKind, scopeKind, permission = permissions.KindPhoneHome, permissions.KindPhoneHome, permissions.PhoneHomeCreate
		verb = permissions.PermissionCreate
	case app.RoleTypeLogStream:
		objectKind, scopeKind, permission = permissions.KindLogStream, permissions.KindLogStream, permissions.LogStreamCreate
		verb = permissions.PermissionCreate
	default:
		return permissions.LegacyPolicy{}, fmt.Errorf("unsupported scoped role type %q", role.RoleType)
	}
	projection, err := role.Policies[0].PermissionProjection()
	if err != nil {
		return permissions.LegacyPolicy{}, err
	}
	api, enforced, err := APIFromRole(role)
	if err != nil {
		return permissions.LegacyPolicy{}, err
	}
	keys := 1
	if enforced {
		keys++
	}
	if len(projection) != keys {
		return permissions.LegacyPolicy{}, fmt.Errorf("scoped role has an unexpected policy shape")
	}
	prefix := permissions.Object(role.OrgID.String, objectKind, "")
	for object, value := range projection {
		if object == permissions.APIPermissionKey {
			continue
		}
		if !strings.HasPrefix(object, prefix) || value == nil || *value != string(verb) {
			return permissions.LegacyPolicy{}, fmt.Errorf("scoped role has an unexpected resource grant")
		}
		return permissions.LegacyResourcePolicy(api, permissions.Scope{
			Kind: scopeKind, OrgID: role.OrgID.String, ID: strings.TrimPrefix(object, prefix),
		}, permission)
	}
	return permissions.LegacyPolicy{}, fmt.Errorf("scoped role requires a resource grant")
}
