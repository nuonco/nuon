package operationroles

import "github.com/nuonco/nuon/services/ctl-api/internal/app"

type EntityOperationRoleMap map[app.OperationType]string

func EntityOperationRoleMapFromHstore(hstore map[string]*string) EntityOperationRoleMap {
	if hstore == nil {
		return nil
	}

	result := make(EntityOperationRoleMap, len(hstore))
	for key, value := range hstore {
		if value != nil {
			result[app.OperationType(key)] = *value
		}
	}
	return result
}

func NewPermissionInfo(rs *RoleSelection) app.RunnerJobPermissionInfo {
	if rs == nil {
		return app.RunnerJobPermissionInfo{}
	}
	return app.RunnerJobPermissionInfo{
		Role:               rs.RoleName,
		RoleSource:         string(rs.Source),
		RoleSelectionTrace: rs.Trace,
	}
}
