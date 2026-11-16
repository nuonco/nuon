package permissions

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	strictjson "sigs.k8s.io/json"
)

const PolicyFormatV4 = 4

type NativeScope struct {
	Kind ResourceKind `json:"kind"`
	IDs  []string     `json:"ids"`
}

type NativeGrant struct {
	Scope       NativeScope    `json:"scope"`
	Permissions []PermissionID `json:"permissions"`
}

type NativePolicy struct {
	Format int           `json:"format"`
	API    API           `json:"api,omitempty"`
	Grants []NativeGrant `json:"grants"`
}

type CompiledPolicy struct {
	orgID    string
	api      API
	revision string
	grants   []compiledGrant
}

type compiledGrant struct {
	scope      NativeScope
	operations []PermissionID
}

func (c *Catalog) DecodePolicy(data []byte) (NativePolicy, error) {
	var document struct {
		Format int             `json:"format"`
		API    json.RawMessage `json:"api,omitempty"`
		Grants []NativeGrant   `json:"grants"`
	}
	strictErrors, err := strictjson.UnmarshalStrict(data, &document)
	if err != nil {
		return NativePolicy{}, err
	}
	if err := errors.Join(strictErrors...); err != nil {
		return NativePolicy{}, err
	}
	policy := NativePolicy{Format: document.Format, Grants: document.Grants}
	if document.API != nil {
		if err := json.Unmarshal(document.API, &policy.API); err != nil {
			return NativePolicy{}, err
		}
		if _, err := ParseAPI(string(policy.API)); err != nil {
			return NativePolicy{}, err
		}
	}
	return c.NormalizePolicy(policy)
}

func (c *Catalog) NormalizePolicy(policy NativePolicy) (NativePolicy, error) {
	if policy.Format != PolicyFormatV4 {
		return NativePolicy{}, fmt.Errorf("unsupported native policy format %d", policy.Format)
	}
	if policy.API != "" {
		if _, err := ParseAPI(string(policy.API)); err != nil {
			return NativePolicy{}, err
		}
	}
	if policy.Grants == nil {
		return NativePolicy{}, fmt.Errorf("native policy requires a grants array")
	}
	normalized := NativePolicy{Format: policy.Format, API: policy.API, Grants: make([]NativeGrant, 0, len(policy.Grants))}
	for _, grant := range policy.Grants {
		if err := grant.Scope.Kind.Validate(); err != nil {
			return NativePolicy{}, err
		}
		if len(grant.Scope.IDs) == 0 {
			return NativePolicy{}, fmt.Errorf("native scope requires a nonempty IDs array")
		}
		for _, id := range grant.Scope.IDs {
			if id == "*" && len(grant.Scope.IDs) == 1 {
				continue
			}
			if !validScopeID(id) {
				return NativePolicy{}, fmt.Errorf("invalid native scope ID %q", id)
			}
		}
		if grant.Permissions == nil {
			return NativePolicy{}, fmt.Errorf("native grant requires a permissions array")
		}
		for _, permission := range grant.Permissions {
			definition, ok := c.definitions[permission]
			if !ok || !slices.Contains(definition.Scopes, grant.Scope.Kind) {
				return NativePolicy{}, fmt.Errorf("unsupported permission %q at %q scope", permission, grant.Scope.Kind)
			}
		}
		normalized.Grants = append(normalized.Grants, NativeGrant{
			Scope:       NativeScope{Kind: grant.Scope.Kind, IDs: sortedUnique(grant.Scope.IDs)},
			Permissions: sortedUnique(grant.Permissions),
		})
	}
	return normalized, nil
}

func (c *Catalog) Compile(policy NativePolicy, orgID string) (*CompiledPolicy, error) {
	if !validScopeID(orgID) {
		return nil, fmt.Errorf("native policy requires an explicit trusted org ID")
	}
	normalized, err := c.NormalizePolicy(policy)
	if err != nil {
		return nil, err
	}
	compiled := &CompiledPolicy{orgID: orgID, api: normalized.API, revision: c.revision}
	for _, grant := range normalized.Grants {
		if grant.Scope.Kind == KindOrg {
			for _, id := range grant.Scope.IDs {
				if id != "*" && id != orgID {
					return nil, fmt.Errorf("org scope must select the trusted org")
				}
			}
		}
		operations := make([]PermissionID, 0)
		for _, reference := range grant.Permissions {
			operations = append(operations, c.expanded[reference]...)
		}
		compiled.grants = append(compiled.grants, compiledGrant{scope: grant.Scope, operations: sortedUnique(operations)})
	}
	return compiled, nil
}
