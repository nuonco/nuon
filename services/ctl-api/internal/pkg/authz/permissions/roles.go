package permissions

import "fmt"

type RoleName string

const (
	RoleOrgReader RoleName = "org_read_only"
	RoleOrgAdmin  RoleName = "org_admin"
)

func LegacyRolePolicy(name RoleName, api API) (LegacyPolicy, error) {
	if api != "" {
		if _, err := ParseAPI(string(api)); err != nil {
			return LegacyPolicy{}, err
		}
	}
	values := []PermissionID{OrgRead}
	switch name {
	case RoleOrgReader:
	case RoleOrgAdmin:
		values = append(values, OrgWrite)
	default:
		return LegacyPolicy{}, fmt.Errorf("invalid role %q", name)
	}
	set, err := NewPermissionSet(values...)
	if err != nil {
		return LegacyPolicy{}, err
	}
	return LegacyPolicy{kind: PolicyKindBundle, api: api, bindingKind: KindOrg, permissions: set}, nil
}
