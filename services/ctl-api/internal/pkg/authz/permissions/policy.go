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

const (
	PolicyKindBundle        PolicyKind = "bundle"
	PolicyKindResourceRules PolicyKind = "resource_rules"
)

type Policy struct {
	api         API
	kind        PolicyKind
	bindingKind ResourceKind
	permissions PermissionSet
	rules       []ScopedPermissions
}

type bundleDocument struct {
	Kind        PolicyKind     `json:"kind"`
	API         API            `json:"api"`
	BindingKind ResourceKind   `json:"binding_kind"`
	Permissions []PermissionID `json:"permissions"`
}

type resourceRule struct {
	Scope       Scope          `json:"scope"`
	Permissions []PermissionID `json:"permissions"`
}

type resourceDocument struct {
	Kind  PolicyKind     `json:"kind"`
	API   API            `json:"api"`
	Rules []resourceRule `json:"rules"`
}

func DecodePolicy(version int, data []byte) (Policy, error) {
	if version != PolicyFormatV3 {
		return Policy{}, fmt.Errorf("unsupported policy format %d", version)
	}
	var header struct {
		Kind PolicyKind `json:"kind"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return Policy{}, err
	}
	var policy Policy
	switch header.Kind {
	case PolicyKindBundle:
		var doc bundleDocument
		if err := decodePolicyDocument(data, &doc); err != nil {
			return Policy{}, err
		}
		if doc.API != APICtl {
			return Policy{}, fmt.Errorf("customer bundle requires ctl-api eligibility")
		}
		if doc.BindingKind != KindOrg && doc.BindingKind != KindApp && doc.BindingKind != KindInstall {
			return Policy{}, fmt.Errorf("invalid bundle binding kind %q", doc.BindingKind)
		}
		if doc.Permissions == nil {
			return Policy{}, fmt.Errorf("bundle requires a permission array")
		}
		set, err := NewPermissionSet(doc.Permissions...)
		if err != nil {
			return Policy{}, err
		}
		for permission := range set {
			if !permission.customerPermission() {
				return Policy{}, fmt.Errorf("permission %q is not a customer permission", permission)
			}
		}
		policy = Policy{kind: doc.Kind, api: doc.API, bindingKind: doc.BindingKind, permissions: set}
	case PolicyKindResourceRules:
		var doc resourceDocument
		if err := decodePolicyDocument(data, &doc); err != nil {
			return Policy{}, err
		}
		if doc.Rules == nil {
			return Policy{}, fmt.Errorf("resource policy requires a rule array")
		}
		policy = Policy{kind: doc.Kind, api: doc.API, rules: make([]ScopedPermissions, 0, len(doc.Rules))}
		for _, rule := range doc.Rules {
			if err := rule.Scope.ValidateShape(); err != nil {
				return Policy{}, err
			}
			if rule.Scope.Kind == KindOrg || rule.Scope.Kind == KindApp || rule.Scope.Kind == KindInstall {
				return Policy{}, fmt.Errorf("workload rule requires an exact resource, not a tenant subtree")
			}
			if rule.Permissions == nil {
				return Policy{}, fmt.Errorf("resource rule requires a permission array")
			}
			set, err := NewPermissionSet(rule.Permissions...)
			if err != nil {
				return Policy{}, err
			}
			for permission := range set {
				if !permission.appliesTo(rule.Scope.Kind) {
					return Policy{}, fmt.Errorf("permission %q does not apply to %q", permission, rule.Scope.Kind)
				}
			}
			policy.rules = append(policy.rules, ScopedPermissions{Scope: rule.Scope, Permissions: set})
		}
	default:
		return Policy{}, fmt.Errorf("invalid policy kind %q", header.Kind)
	}
	if _, err := ParseAPI(string(policy.api)); err != nil {
		return Policy{}, err
	}
	return policy, nil
}

func decodePolicyDocument(data []byte, value any) error {
	strictErrors, err := strictjson.UnmarshalStrict(data, value)
	if err != nil {
		return err
	}
	return errors.Join(strictErrors...)
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
	switch p.kind {
	case PolicyKindBundle:
		if binding.Kind != p.bindingKind {
			return nil, fmt.Errorf("bundle requires binding kind %q", p.bindingKind)
		}
		return []ScopedPermissions{{Scope: binding, Permissions: maps.Clone(p.permissions)}}, nil
	case PolicyKindResourceRules:
		if binding.Kind != KindOrg {
			return nil, fmt.Errorf("private resource policy requires an org association")
		}
		grants := make([]ScopedPermissions, 0, len(p.rules))
		for _, rule := range p.rules {
			if rule.Scope.OrgID != orgID {
				return nil, fmt.Errorf("resource rule is outside the policy owner's tenant")
			}
			grants = append(grants, ScopedPermissions{Scope: rule.Scope, Permissions: maps.Clone(rule.Permissions)})
		}
		return grants, nil
	default:
		return nil, fmt.Errorf("invalid policy kind %q", p.kind)
	}
}

func (p Policy) MarshalJSON() ([]byte, error) {
	switch p.kind {
	case PolicyKindBundle:
		return json.Marshal(bundleDocument{
			Kind: p.kind, API: p.api, BindingKind: p.bindingKind, Permissions: p.permissions.Values(),
		})
	case PolicyKindResourceRules:
		rules := make([]resourceRule, 0, len(p.rules))
		for _, rule := range p.rules {
			rules = append(rules, resourceRule{Scope: rule.Scope, Permissions: rule.Permissions.Values()})
		}
		return json.Marshal(resourceDocument{Kind: p.kind, API: p.api, Rules: rules})
	default:
		return nil, fmt.Errorf("invalid policy kind %q", p.kind)
	}
}
