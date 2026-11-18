package permissions

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	strictjson "sigs.k8s.io/json"
)

const PolicyFormatV4 = 4

type GrantScope struct {
	Kind ResourceKind `json:"kind"`
	IDs  []string     `json:"ids"`
}

type Grant struct {
	Scope       GrantScope     `json:"scope"`
	Permissions []PermissionID `json:"permissions"`
}

type Policy struct {
	Format int     `json:"format"`
	API    API     `json:"api,omitempty"`
	Grants []Grant `json:"grants"`
}

type CompiledPolicy struct {
	orgID    string
	api      API
	revision string
	grants   []compiledGrant
}

type compiledGrant struct {
	scope      GrantScope
	operations []PermissionID
}

func (c *Catalog) DecodePolicy(data []byte) (Policy, error) {
	var document struct {
		Format int             `json:"format"`
		API    json.RawMessage `json:"api,omitempty"`
		Grants []Grant         `json:"grants"`
	}
	strictErrors, err := strictjson.UnmarshalStrict(data, &document)
	if err != nil {
		return Policy{}, err
	}
	if err := errors.Join(strictErrors...); err != nil {
		return Policy{}, err
	}
	policy := Policy{Format: document.Format, Grants: document.Grants}
	if document.API != nil {
		if err := json.Unmarshal(document.API, &policy.API); err != nil {
			return Policy{}, err
		}
		if _, err := ParseAPI(string(policy.API)); err != nil {
			return Policy{}, err
		}
	}
	return c.NormalizePolicy(policy)
}

func (c *Catalog) NormalizePolicy(policy Policy) (Policy, error) {
	if policy.Format != PolicyFormatV4 {
		return Policy{}, fmt.Errorf("unsupported grant policy format %d", policy.Format)
	}
	if policy.API != "" {
		if _, err := ParseAPI(string(policy.API)); err != nil {
			return Policy{}, err
		}
	}
	if policy.Grants == nil {
		return Policy{}, fmt.Errorf("policy requires a grants array")
	}
	normalized := Policy{Format: policy.Format, API: policy.API, Grants: make([]Grant, 0, len(policy.Grants))}
	for _, grant := range policy.Grants {
		if err := grant.Scope.Kind.Validate(); err != nil {
			return Policy{}, err
		}
		if len(grant.Scope.IDs) == 0 {
			return Policy{}, fmt.Errorf("grant scope requires a nonempty IDs array")
		}
		for _, id := range grant.Scope.IDs {
			if id == "*" && len(grant.Scope.IDs) == 1 {
				continue
			}
			if !validScopeID(id) {
				return Policy{}, fmt.Errorf("invalid grant scope ID %q", id)
			}
		}
		if grant.Permissions == nil {
			return Policy{}, fmt.Errorf("grant requires a permissions array")
		}
		for _, permission := range grant.Permissions {
			definition, ok := c.definitions[permission]
			if !ok || !slices.Contains(definition.Scopes, grant.Scope.Kind) {
				return Policy{}, fmt.Errorf("unsupported permission %q at %q scope", permission, grant.Scope.Kind)
			}
		}
		normalized.Grants = append(normalized.Grants, Grant{
			Scope:       GrantScope{Kind: grant.Scope.Kind, IDs: sortedUnique(grant.Scope.IDs)},
			Permissions: sortedUnique(grant.Permissions),
		})
	}
	return normalized, nil
}

func (c *Catalog) Compile(policy Policy, orgID string) (*CompiledPolicy, error) {
	if !validScopeID(orgID) {
		return nil, fmt.Errorf("policy requires an explicit trusted org ID")
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
