package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidc"
)

func oidcPatternValidator(fl validator.FieldLevel) bool {
	_, err := oidc.Compile(fl.Field().String())
	return err == nil
}

func oidcClaimNameValidator(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	return name != "" && !oidc.IsReservedClaim(name)
}

// oidcFieldError formats errors on oidc.Conditions, whose field namespaces
// identify the claim: `Conditions.Sub` or `Conditions.Extra[<claim>]`.
func oidcFieldError(e validator.FieldError) (string, bool) {
	if !strings.HasPrefix(e.Namespace(), oidc.ConditionsStruct+".") {
		return "", false
	}

	claim := "sub"
	if e.Field() != "Sub" {
		claim = strings.TrimSuffix(strings.TrimPrefix(e.Field(), "Extra["), "]")
	}

	switch e.Tag() {
	case "required":
		if claim == "sub" {
			return "a condition on the `sub` claim is required", true
		}
		return fmt.Sprintf("condition for claim %q cannot be empty", claim), true
	case "max":
		return fmt.Sprintf("condition for claim %q exceeds %s characters", claim, e.Param()), true
	case "oidc_pattern":
		return fmt.Sprintf("condition for claim %q is not a valid pattern", claim), true
	case "oidc_claim_name":
		return fmt.Sprintf("claim %q cannot be used as a condition", claim), true
	}

	return "", false
}
