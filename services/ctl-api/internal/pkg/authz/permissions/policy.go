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

type policyHeader struct {
	Kind        PolicyKind      `json:"kind"`
	API         json.RawMessage `json:"api,omitempty"`
	BindingKind ResourceKind    `json:"binding_kind"`
}

type bundleDocument struct {
	policyHeader
	Permissions []PermissionID `json:"permissions"`
}

type resourceScope struct {
	Kind ResourceKind `json:"kind"`
	ID   string       `json:"id"`
}

type resourceRule struct {
	Scope       resourceScope  `json:"scope"`
	Permissions []PermissionID `json:"permissions"`
}

type resourceDocument struct {
	policyHeader
	Rules []resourceRule `json:"rules"`
}

func ResourcePolicy(api API, scope Scope, permission PermissionID) (Policy, error) {
	if api != "" {
		if _, err := ParseAPI(string(api)); err != nil {
			return Policy{}, err
		}
	}
	switch scope.Kind {
	case KindInstall, KindRunnerJob, KindPhoneHome, KindLogStream:
	default:
		return Policy{}, fmt.Errorf("unsupported resource rule scope %q", scope.Kind)
	}
	set, err := NewPermissionSet(permission)
	if err != nil {
		return Policy{}, err
	}
	rules := []ScopedPermissions{{Scope: scope, Permissions: set}}
	if _, err := LegacySources(rules); err != nil {
		return Policy{}, err
	}
	return Policy{kind: PolicyKindResourceRules, api: api, bindingKind: KindOrg, rules: rules}, nil
}

func DecodePolicy(version int, data []byte) (Policy, error) {
	if version != PolicyFormatV3 {
		return Policy{}, fmt.Errorf("unsupported policy format %d", version)
	}
	var header policyHeader
	if err := json.Unmarshal(data, &header); err != nil {
		return Policy{}, err
	}
	var bundle bundleDocument
	var resources resourceDocument
	var doc any
	switch header.Kind {
	case PolicyKindBundle:
		doc = &bundle
	case PolicyKindResourceRules:
		doc = &resources
	default:
		return Policy{}, fmt.Errorf("invalid policy kind %q", header.Kind)
	}
	strictErrors, err := strictjson.UnmarshalStrict(data, doc)
	if err != nil {
		return Policy{}, err
	}
	if err := errors.Join(strictErrors...); err != nil {
		return Policy{}, err
	}
	var api API
	if header.API != nil {
		if err := json.Unmarshal(header.API, &api); err != nil {
			return Policy{}, err
		}
		if _, err := ParseAPI(string(api)); err != nil {
			return Policy{}, err
		}
	}
	if header.BindingKind != KindOrg {
		return Policy{}, fmt.Errorf("invalid policy binding kind %q", header.BindingKind)
	}
	policy := Policy{kind: header.Kind, api: api, bindingKind: header.BindingKind}
	if header.Kind == PolicyKindBundle {
		if bundle.Permissions == nil {
			return Policy{}, fmt.Errorf("bundle requires a permission array")
		}
		policy.permissions, err = NewPermissionSet(bundle.Permissions...)
		return policy, err
	}
	if resources.Rules == nil {
		return Policy{}, fmt.Errorf("resource policy requires a rule array")
	}
	for _, rule := range resources.Rules {
		if !validScopeID(rule.Scope.ID) {
			return Policy{}, fmt.Errorf("resource rule requires an explicit resource ID")
		}
		switch rule.Scope.Kind {
		case KindInstall, KindRunnerJob, KindPhoneHome, KindLogStream:
		default:
			return Policy{}, fmt.Errorf("unsupported resource rule scope %q", rule.Scope.Kind)
		}
		if rule.Permissions == nil {
			return Policy{}, fmt.Errorf("resource rule requires a permission array")
		}
		set, err := NewPermissionSet(rule.Permissions...)
		if err != nil {
			return Policy{}, err
		}
		policy.rules = append(policy.rules, ScopedPermissions{
			Scope: Scope{Kind: rule.Scope.Kind, ID: rule.Scope.ID}, Permissions: set,
		})
	}
	return policy, nil
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
	if binding.Kind != p.bindingKind {
		return nil, fmt.Errorf("policy requires binding kind %q", p.bindingKind)
	}
	switch p.kind {
	case PolicyKindBundle:
		return []ScopedPermissions{{Scope: binding, Permissions: maps.Clone(p.permissions)}}, nil
	case PolicyKindResourceRules:
		grants := make([]ScopedPermissions, 0, len(p.rules))
		for _, rule := range p.rules {
			scope := rule.Scope
			scope.OrgID = orgID
			if err := scope.ValidateShape(); err != nil {
				return nil, err
			}
			grants = append(grants, ScopedPermissions{Scope: scope, Permissions: maps.Clone(rule.Permissions)})
		}
		return grants, nil
	default:
		return nil, fmt.Errorf("invalid policy kind %q", p.kind)
	}
}

func (p Policy) MarshalJSON() ([]byte, error) {
	var api json.RawMessage
	if p.api != "" {
		value, err := json.Marshal(p.api)
		if err != nil {
			return nil, err
		}
		api = value
	}
	header := policyHeader{Kind: p.kind, API: api, BindingKind: p.bindingKind}
	switch p.kind {
	case PolicyKindBundle:
		return json.Marshal(bundleDocument{policyHeader: header, Permissions: p.permissions.Values()})
	case PolicyKindResourceRules:
		rules := make([]resourceRule, 0, len(p.rules))
		for _, rule := range p.rules {
			rules = append(rules, resourceRule{
				Scope: resourceScope{Kind: rule.Scope.Kind, ID: rule.Scope.ID}, Permissions: rule.Permissions.Values(),
			})
		}
		return json.Marshal(resourceDocument{policyHeader: header, Rules: rules})
	default:
		return nil, fmt.Errorf("invalid policy kind %q", p.kind)
	}
}
