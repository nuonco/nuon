package permissions

import (
	"fmt"

	"github.com/nuonco/nuon/pkg/generics"
)

type Permission string

const (
	PermissionUnknown Permission = "unknown"

	PermissionAll    Permission = "all"
	PermissionCreate Permission = "create"
	PermissionRead   Permission = "read"
	PermissionUpdate Permission = "update"
	PermissionDelete Permission = "delete"
)

func (p Permission) ToStrPtr() *string {
	return generics.ToPtr(string(p))
}

func NewVerbs(vals []string) (Verbs, error) {
	if len(vals) == 0 {
		return nil, fmt.Errorf("at least one permission is required")
	}

	var out Verbs
	for _, val := range vals {
		perm, err := NewPermission(val)
		if err != nil {
			return nil, err
		}
		out = out.With(perm)
	}
	return out, nil
}

func NewPermission(val string) (Permission, error) {
	switch val {
	case "all":
		return PermissionAll, nil
	case "create":
		return PermissionCreate, nil
	case "update":
		return PermissionUpdate, nil
	case "read":
		return PermissionRead, nil
	case "delete":
		return PermissionDelete, nil
	default:
	}

	return PermissionUnknown, fmt.Errorf("invalid permission %s", val)
}
