package permissions

import (
	"encoding/json"
	"fmt"
)

type CredentialProfile string

const (
	ProfileHuman          CredentialProfile = "human"
	ProfileServiceAccount CredentialProfile = "service-account"
)

type CredentialLimits struct {
	Profile CredentialProfile
	API     API
	Ceiling []ScopedPermissions
}

type credentialDocument struct {
	Version int               `json:"version"`
	Profile CredentialProfile `json:"profile"`
	API     API               `json:"api"`
	Ceiling []resourceRule    `json:"ceiling"`
}

func DecodeCredentialLimits(data []byte) (CredentialLimits, error) {
	var doc credentialDocument
	if err := decodePolicyDocument(data, &doc); err != nil {
		return CredentialLimits{}, err
	}
	if doc.Version != 1 || doc.Ceiling == nil {
		return CredentialLimits{}, fmt.Errorf("unsupported credential restrictions")
	}
	limits := CredentialLimits{Profile: doc.Profile, API: doc.API, Ceiling: make([]ScopedPermissions, 0, len(doc.Ceiling))}
	for _, rule := range doc.Ceiling {
		if rule.Permissions == nil {
			return CredentialLimits{}, fmt.Errorf("credential scope requires explicit permissions")
		}
		set, err := NewPermissionSet(rule.Permissions...)
		if err != nil {
			return CredentialLimits{}, err
		}
		limits.Ceiling = append(limits.Ceiling, ScopedPermissions{Scope: rule.Scope, Permissions: set})
	}
	return limits, limits.Validate()
}

func (l CredentialLimits) Validate() error {
	if l.Profile != ProfileHuman && l.Profile != ProfileServiceAccount {
		return fmt.Errorf("unsupported credential profile %q", l.Profile)
	}
	if _, err := ParseAPI(string(l.API)); err != nil {
		return err
	}
	if l.Ceiling == nil || !validGrants(l.Ceiling) {
		return fmt.Errorf("credential requires an explicit valid ceiling")
	}
	return nil
}

func (l CredentialLimits) MarshalJSON() ([]byte, error) {
	if err := l.Validate(); err != nil {
		return nil, err
	}
	doc := credentialDocument{Version: 1, Profile: l.Profile, API: l.API, Ceiling: make([]resourceRule, 0, len(l.Ceiling))}
	for _, grant := range l.Ceiling {
		doc.Ceiling = append(doc.Ceiling, resourceRule{Scope: grant.Scope, Permissions: grant.Permissions.Values()})
	}
	return json.Marshal(doc)
}
