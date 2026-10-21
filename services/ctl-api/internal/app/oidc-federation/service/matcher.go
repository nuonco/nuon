package service

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidc"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidc/github"
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
		if _, err := oidc.Compile(pattern); err != nil {
			return fmt.Errorf("condition for claim %q is not a valid pattern: %w", claim, err)
		}
	}

	sub := conditions["sub"]
	isGitHub := github.IsIssuer(issuerURL)

	if isGitHub {
		if err := validateGitHubOwnerPinned(conditions); err != nil {
			return err
		}
	}

	if subPatternTooBroad(sub, isGitHub) {
		return fmt.Errorf("the `sub` condition is too broad: it matches unrelated identities. " +
			"Put wildcards only after the part that identifies your repository or workload")
	}

	return nil
}

// validateGitHubOwnerPinned requires either a literal owner prefix in the
// `sub` pattern or a literal numeric `repository_owner_id` condition.
func validateGitHubOwnerPinned(conditions map[string]string) error {
	if _, ok := github.ParseSubjectOwner(conditions["sub"]); ok {
		return nil
	}

	if isDigits(conditions["repository_owner_id"]) {
		return nil
	}

	return fmt.Errorf("GitHub Actions tokens can come from any account, so the `sub` must pin an owner: " +
		"start it with a literal owner like `repo:acme/`, or add a numeric `repository_owner_id` condition")
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
func subPatternTooBroad(pattern string, isGitHub bool) bool {
	a, b, c, d := probeA, probeB, probeC, probeD

	probes := []string{
		a,
		a + ":" + b,
		a + ":" + b + ":" + c,
		a + ":" + b + ":" + c + ":" + d,
		a + "/" + b,
	}
	if isGitHub {
		repo := github.NewSubject(a, b, "")
		probes = append(probes,
			repo.WithContext("ref:refs/heads/main").String(),
			repo.WithContext("environment:production").String(),
			repo.WithContext("pull_request").String(),
			repo.WithIDs("1", "2").WithContext("ref:refs/heads/main").String(),
		)
	}

	for _, probe := range probes {
		if oidc.MatchPattern(pattern, probe) {
			return true
		}
	}

	return false
}
