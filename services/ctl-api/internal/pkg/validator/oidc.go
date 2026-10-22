package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidc"
)

func OIDCConditions(v *validator.Validate, providers oidc.Providers, issuerURL string, conditions map[string]string) error {
	c := oidc.ParseConditions(conditions)

	if err := v.Struct(c); err != nil {
		return oidcConditionsError(err)
	}

	provider := providers.For(issuerURL)

	if err := provider.ValidateConditions(v, c); err != nil {
		return err
	}

	if oidc.TooBroad(c.Sub, provider.Probes()) {
		return oidc.ErrTooBroad
	}

	return nil
}

func oidcPatternValidator(fl validator.FieldLevel) bool {
	_, err := oidc.Compile(fl.Field().String())
	return err == nil
}

func oidcClaimNameValidator(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	return name != "" && !oidc.IsReservedClaim(name)
}

func oidcConditionsError(err error) error {
	var errs validator.ValidationErrors
	if !errors.As(err, &errs) {
		return err
	}

	msgs := make([]string, 0, len(errs))
	for _, e := range errs {
		msgs = append(msgs, oidcFieldError(e))
	}

	return errors.New(strings.Join(msgs, "; "))
}

func oidcFieldError(e validator.FieldError) string {
	claim := "sub"
	if e.Field() != "Sub" {
		claim = strings.TrimSuffix(strings.TrimPrefix(e.Field(), "Extra["), "]")
	}

	switch e.Tag() {
	case "required":
		if claim == "sub" {
			return "a condition on the `sub` claim is required"
		}
		return fmt.Sprintf("condition for claim %q cannot be empty", claim)
	case "max":
		return fmt.Sprintf("condition for claim %q exceeds %s characters", claim, e.Param())
	case "oidc_pattern":
		return fmt.Sprintf("condition for claim %q is not a valid pattern", claim)
	case "oidc_claim_name":
		return fmt.Sprintf("claim %q cannot be used as a condition", claim)
	}

	return fmt.Sprintf("condition for claim %q is invalid", claim)
}
