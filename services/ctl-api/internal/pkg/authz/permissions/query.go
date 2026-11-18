package permissions

import (
	"fmt"
	"slices"
)

type Query struct {
	API        API
	Permission PermissionID
	Kind       ResourceKind
	OrgID      string
	Grants     []*CompiledPolicy
	Ceiling    []*CompiledPolicy
}

func (c *Catalog) QueryScopes(input Query) (grants, ceiling []GrantScope, err error) {
	if _, err := ParseAPI(string(input.API)); err != nil {
		return nil, nil, err
	}
	if !validScopeID(input.OrgID) {
		return nil, nil, fmt.Errorf("permission query requires a trusted org ID")
	}
	definition, ok := c.definitions[input.Permission]
	if !ok || len(definition.Members) != 0 || !slices.Contains(definition.Targets, input.Kind) {
		return nil, nil, fmt.Errorf("unsupported query operation %q for %q", input.Permission, input.Kind)
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

func (c *Catalog) operationScopes(policies []*CompiledPolicy, orgID string, api API, operation PermissionID) ([]GrantScope, error) {
	result := make([]GrantScope, 0)
	for _, policy := range policies {
		if policy == nil || policy.revision != c.revision {
			return nil, fmt.Errorf("permissions require policies compiled with the current catalog")
		}
		if policy.orgID != orgID || (policy.api != "" && policy.api != api) {
			continue
		}
		for _, grant := range policy.grants {
			if slices.Contains(grant.operations, operation) {
				result = append(result, GrantScope{Kind: grant.scope.Kind, IDs: slices.Clone(grant.scope.IDs)})
			}
		}
	}
	return result, nil
}
