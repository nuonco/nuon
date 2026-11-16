package permissions

import (
	"fmt"
	"slices"
)

type NativePermissionQuery struct {
	API        API
	Permission PermissionID
	Kind       ResourceKind
	OrgID      string
	Grants     []*CompiledPolicy
	Ceiling    []*CompiledPolicy
}

func (c *Catalog) QueryScopes(input NativePermissionQuery) (grants, ceiling []NativeScope, err error) {
	if _, err := ParseAPI(string(input.API)); err != nil {
		return nil, nil, err
	}
	if !validScopeID(input.OrgID) {
		return nil, nil, fmt.Errorf("native query requires a trusted org ID")
	}
	definition, ok := c.definitions[input.Permission]
	if !ok || len(definition.Members) != 0 || !slices.Contains(definition.Targets, input.Kind) {
		return nil, nil, fmt.Errorf("unsupported native query operation %q for %q", input.Permission, input.Kind)
	}
	grants, err = c.operationScopes(input.Grants, input.OrgID, input.API, input.Permission)
	if err != nil {
		return nil, nil, err
	}
	if input.Ceiling != nil {
		ceiling, err = c.operationScopes(input.Ceiling, input.OrgID, input.API, input.Permission)
		if err != nil {
			return nil, nil, err
		}
	}
	return grants, ceiling, nil
}

func (c *Catalog) operationScopes(policies []*CompiledPolicy, orgID string, api API, operation PermissionID) ([]NativeScope, error) {
	result := make([]NativeScope, 0)
	for _, policy := range policies {
		if policy == nil || policy.revision != c.revision {
			return nil, fmt.Errorf("native permissions require policies compiled with the current catalog")
		}
		if policy.orgID != orgID || (policy.api != "" && policy.api != api) {
			continue
		}
		for _, grant := range policy.grants {
			if slices.Contains(grant.operations, operation) {
				result = append(result, NativeScope{Kind: grant.scope.Kind, IDs: slices.Clone(grant.scope.IDs)})
			}
		}
	}
	return result, nil
}
