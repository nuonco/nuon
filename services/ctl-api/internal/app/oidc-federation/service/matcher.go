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
//
// issuerURL is the policy's effective issuer. Multi-tenant issuers (GitHub
// Actions) additionally require the policy to be pinned to a specific owner,
// and every issuer is checked against a backstop that rejects `sub` patterns
// matching random foreign values.
func validateClaimConditions(issuerURL string, conditions map[string]string) error {
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

	sub := conditions["sub"]
	github := isGitHubActionsIssuer(issuerURL)

	if github {
		if err := validateGitHubOwnerPinned(conditions); err != nil {
			return err
		}
	}

	if subPatternTooBroad(sub, github) {
		return fmt.Errorf("the `sub` condition is too broad: it matches unrelated identities. " +
			"Put wildcards only after the part that identifies your repository or workload")
	}

	return nil
}

// githubActionsIssuer is GitHub's multi-tenant OIDC issuer: any GitHub user
// can mint tokens from it, so policies must be pinned to an owner.
const githubActionsIssuer = "https://token.actions.githubusercontent.com"

const globMetaChars = "*?[]{}\\"

func isGitHubActionsIssuer(issuerURL string) bool {
	return strings.TrimSuffix(issuerURL, "/") == githubActionsIssuer
}

// validateGitHubOwnerPinned requires either a literal owner prefix in the
// `sub` pattern or a literal numeric `repository_owner_id` condition.
func validateGitHubOwnerPinned(conditions map[string]string) error {
	if subPinsGitHubOwner(conditions["sub"]) {
		return nil
	}

	if isDigits(conditions["repository_owner_id"]) {
		return nil
	}

	return fmt.Errorf("GitHub Actions tokens can come from any account, so the `sub` must pin an owner: " +
		"start it with a literal owner like `repo:acme/`, or add a numeric `repository_owner_id` condition")
}

// subPinsGitHubOwner reports whether pattern starts with `repo:<owner>/` where
// the owner is non-empty and free of glob metacharacters.
func subPinsGitHubOwner(pattern string) bool {
	rest, ok := strings.CutPrefix(pattern, "repo:")
	if !ok {
		return false
	}

	owner, _, ok := strings.Cut(rest, "/")
	if !ok || owner == "" {
		return false
	}

	return !strings.ContainsAny(owner, globMetaChars+":")
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}

// Probe values stand in for identities a policy author has never heard of.
// They are fixed rather than random so validation is deterministic.
const (
	probeA = "nuonprobeqz7a"
	probeB = "nuonprobekx3b"
	probeC = "nuonprobewv5c"
	probeD = "nuonprobejm9d"
)

// subPatternTooBroad reports whether the `sub` pattern matches foreign probe
// values, which no policy for a specific identity should.
func subPatternTooBroad(pattern string, github bool) bool {
	a, b, c, d := probeA, probeB, probeC, probeD

	probes := []string{
		a,
		a + ":" + b,
		a + ":" + b + ":" + c,
		a + ":" + b + ":" + c + ":" + d,
		a + "/" + b,
	}
	if github {
		probes = append(probes,
			"repo:"+a+"/"+b+":ref:refs/heads/main",
			"repo:"+a+"/"+b+":environment:production",
			"repo:"+a+"/"+b+":pull_request",
			"repo:"+a+"@1/"+b+"@2:ref:refs/heads/main",
		)
	}

	for _, probe := range probes {
		if matchPattern(pattern, probe) {
			return true
		}
	}

	return false
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
