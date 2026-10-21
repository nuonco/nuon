package oidc

import (
	"github.com/go-playground/validator/v10"
)

// ValidateConditions enforces the create/update-time rules for a policy's
// claim conditions: `sub` is required so a policy can never match arbitrary
// tokens from an issuer, other claims are not reserved, patterns are bounded
// and compile, the issuer's provider rules hold, and `sub` does not match
// unrelated identities.
//
// Shape errors are validator.ValidationErrors (format them with
// pkg/validator.FormatValidationError); provider and scope errors are plain
// errors. The custom tags are registered by pkg/validator.New.
func ValidateConditions(v *validator.Validate, providers Providers, issuerURL string, conditions map[string]string) error {
	c := ParseConditions(conditions)

	if err := v.Struct(c); err != nil {
		return err
	}

	provider := providers.For(issuerURL)

	if err := provider.ValidateConditions(v, c); err != nil {
		return err
	}

	if TooBroad(c.Sub, provider.Probes()) {
		return ErrTooBroad
	}

	return nil
}
