package permissions

import "fmt"

func LegacyOrgPermission(set PermissionSet) (Permission, error) {
	if len(set) == 0 {
		return PermissionUnknown, nil
	}
	if _, ok := set[OrgRead]; ok {
		if len(set) == 1 {
			return PermissionRead, nil
		}
		if _, ok := set[OrgWrite]; ok && len(set) == 2 {
			return PermissionAll, nil
		}
	}
	return PermissionUnknown, fmt.Errorf("bundle is not supported by the legacy operation adapter")
}
