package permissions

import "fmt"

type RoleName string

const (
	RoleOrgReader RoleName = "org_read_only"
	RoleOrgAdmin  RoleName = "org_admin"
)

func RolePolicy(name RoleName, api API) (Policy, error) {
	if api != "" {
		if _, err := ParseAPI(string(api)); err != nil {
			return Policy{}, err
		}
	}
	values := []PermissionID{OrgRead}
	switch name {
	case RoleOrgReader:
	case RoleOrgAdmin:
		values = append(values, OrgWrite)
	default:
		return Policy{}, fmt.Errorf("invalid role %q", name)
	}
	set, err := NewPermissionSet(values...)
	if err != nil {
		return Policy{}, err
	}
	return Policy{kind: PolicyKindBundle, api: api, bindingKind: KindOrg, permissions: set}, nil
}
