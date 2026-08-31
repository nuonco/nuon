package signature

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/invopop/jsonschema"
)

type AuthorityType string

const (
	AuthorityTypeKeyless   AuthorityType = "keyless"
	AuthorityTypePublicKey AuthorityType = "public_key"
)

type Verification struct {
	RequireSignature bool        `json:"require_signature" mapstructure:"require_signature" toml:"require_signature"`
	Authorities      []Authority `json:"authorities" mapstructure:"authorities" toml:"authorities"`
}

type Authority struct {
	Type          AuthorityType `json:"type" mapstructure:"type" toml:"type" jsonschema:"required,enum=keyless,enum=public_key"`
	Issuer        string        `json:"issuer,omitempty" mapstructure:"issuer,omitempty" toml:"issuer,omitempty"`
	Subject       string        `json:"subject,omitempty" mapstructure:"subject,omitempty" toml:"subject,omitempty"`
	SubjectRegexp string        `json:"subject_regexp,omitempty" mapstructure:"subject_regexp,omitempty" toml:"subject_regexp,omitempty"`
	PublicKey     string        `json:"public_key,omitempty" mapstructure:"public_key,omitempty" toml:"public_key,omitempty" features:"get"`
}

func (Verification) JSONSchemaExtend(_ *jsonschema.Schema) {}

func (Authority) JSONSchemaExtend(_ *jsonschema.Schema) {}

func (v *Verification) Validate() error {
	if v == nil || !v.RequireSignature {
		return nil
	}
	if len(v.Authorities) == 0 {
		return errors.New("verification.authorities must contain at least one authority when require_signature is true")
	}

	for i, authority := range v.Authorities {
		if err := authority.validate(); err != nil {
			return fmt.Errorf("verification.authorities[%d]: %w", i, err)
		}
	}
	return nil
}

func (a Authority) validate() error {
	switch a.Type {
	case AuthorityTypeKeyless:
		if a.Issuer == "" {
			return errors.New("issuer is required for keyless")
		}
		if (a.Subject == "") == (a.SubjectRegexp == "") {
			return errors.New("exactly one of subject or subject_regexp is required for keyless")
		}
		if a.PublicKey != "" {
			return errors.New("public_key is not valid for keyless")
		}
		if a.SubjectRegexp != "" {
			if _, err := regexp.Compile(a.SubjectRegexp); err != nil {
				return fmt.Errorf("invalid subject_regexp: %w", err)
			}
		}
	case AuthorityTypePublicKey:
		if a.PublicKey == "" {
			return errors.New("public_key is required for a public_key authority")
		}
		if a.Issuer != "" || a.Subject != "" || a.SubjectRegexp != "" {
			return errors.New("issuer, subject, and subject_regexp are not valid for a public_key authority")
		}
	default:
		return fmt.Errorf("unsupported type %q", a.Type)
	}
	return nil
}
