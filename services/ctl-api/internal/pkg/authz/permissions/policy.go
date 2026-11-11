package permissions

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"

	strictjson "sigs.k8s.io/json"
)

const PolicyFormatV3 = 3

type PolicyKind string

const PolicyKindBundle PolicyKind = "bundle"

type Policy struct {
	api         API
	kind        PolicyKind
	bindingKind ResourceKind
	permissions PermissionSet
}

type bundleDocument struct {
	Kind        PolicyKind     `json:"kind"`
	API         API            `json:"api"`
	BindingKind ResourceKind   `json:"binding_kind"`
	Permissions []PermissionID `json:"permissions"`
}

func DecodePolicy(version int, data []byte) (Policy, error) {
	if version != PolicyFormatV3 {
		return Policy{}, fmt.Errorf("unsupported policy format %d", version)
	}
	var doc bundleDocument
	strictErrors, err := strictjson.UnmarshalStrict(data, &doc)
	if err != nil {
		return Policy{}, err
	}
	if err := errors.Join(strictErrors...); err != nil {
		return Policy{}, err
	}
	if doc.Kind != PolicyKindBundle {
		return Policy{}, fmt.Errorf("invalid policy kind %q", doc.Kind)
	}
	if _, err := ParseAPI(string(doc.API)); err != nil {
		return Policy{}, err
	}
	if doc.BindingKind != KindOrg {
		return Policy{}, fmt.Errorf("invalid bundle binding kind %q", doc.BindingKind)
	}
	if doc.Permissions == nil {
		return Policy{}, fmt.Errorf("bundle requires a permission array")
	}
	set, err := NewPermissionSet(doc.Permissions...)
	if err != nil {
		return Policy{}, err
	}
	return Policy{kind: doc.Kind, api: doc.API, bindingKind: doc.BindingKind, permissions: set}, nil
}

func (p Policy) API() API {
	return p.api
}

func (p Policy) Kind() PolicyKind {
	return p.kind
}

func (p Policy) Compile(orgID string, binding Scope) ([]ScopedPermissions, error) {
	if !validScopeID(orgID) || binding.OrgID != orgID {
		return nil, fmt.Errorf("policy owner and binding must have the same explicit tenant")
	}
	if err := binding.ValidateShape(); err != nil {
		return nil, err
	}
	if p.kind != PolicyKindBundle {
		return nil, fmt.Errorf("invalid policy kind %q", p.kind)
	}
	if binding.Kind != p.bindingKind {
		return nil, fmt.Errorf("bundle requires binding kind %q", p.bindingKind)
	}
	return []ScopedPermissions{{Scope: binding, Permissions: maps.Clone(p.permissions)}}, nil
}

func (p Policy) MarshalJSON() ([]byte, error) {
	if p.kind != PolicyKindBundle {
		return nil, fmt.Errorf("invalid policy kind %q", p.kind)
	}
	return json.Marshal(bundleDocument{
		Kind: p.kind, API: p.api, BindingKind: p.bindingKind, Permissions: p.permissions.Values(),
	})
}
