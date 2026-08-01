package service

import (
	"fmt"
	"strings"

	"github.com/gobwas/glob"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// reservedClaims are validated by the JWT verifier itself and may not be used
// as claim conditions.
var reservedClaims = map[string]struct{}{
	"iss": {},
	"aud": {},
	"exp": {},
	"nbf": {},
	"iat": {},
}

// validateClaimConditions enforces the create/update-time rules for a
// policy's claim conditions: at least one condition, `sub` must be present so
// a policy can never match arbitrary tokens from an issuer, no reserved
// claims, and bounded non-empty patterns that compile.
func validateClaimConditions(conditions map[string]string) error {
	if len(conditions) == 0 {
		return fmt.Errorf("at least one claim condition is required")
	}

	if _, ok := conditions["sub"]; !ok {
		return fmt.Errorf("a condition on the `sub` claim is required")
	}

	for claim, pattern := range conditions {
		if claim == "" {
			return fmt.Errorf("claim names cannot be empty")
		}
		if _, ok := reservedClaims[claim]; ok {
			return fmt.Errorf("claim %q is validated automatically and cannot be used as a condition", claim)
		}
		if pattern == "" {
			return fmt.Errorf("condition for claim %q cannot be empty", claim)
		}
		if len(pattern) > app.OIDCTrustPolicyMaxPatternLength {
			return fmt.Errorf("condition for claim %q exceeds %d characters", claim, app.OIDCTrustPolicyMaxPatternLength)
		}
		if _, err := compilePattern(pattern); err != nil {
			return fmt.Errorf("condition for claim %q is not a valid pattern: %w", claim, err)
		}
	}

	return nil
}

// matchClaims reports whether every condition matches the corresponding token
// claim. Conditions only match string-typed claims; a missing or non-string
// claim fails the policy.
func matchClaims(conditions map[string]string, claims map[string]any) bool {
	if len(conditions) == 0 {
		return false
	}

	for claim, pattern := range conditions {
		raw, ok := claims[claim]
		if !ok {
			return false
		}

		value, ok := raw.(string)
		if !ok {
			return false
		}

		if !matchPattern(pattern, value) {
			return false
		}
	}

	return true
}

// matchPattern matches a value against a pattern: exact comparison unless the
// pattern contains glob metacharacters, in which case it is compiled with `:`
// as a separator so wildcards cannot cross `:` segments (the delimiter in
// GitHub Actions `sub` claims like `repo:org/repo:ref:refs/heads/main`).
func matchPattern(pattern, value string) bool {
	if !strings.ContainsAny(pattern, "*?[{") {
		return pattern == value
	}

	g, err := compilePattern(pattern)
	if err != nil {
		return false
	}

	return g.Match(value)
}

func compilePattern(pattern string) (glob.Glob, error) {
	return glob.Compile(pattern, ':')
}
