package oidc

// MatchClaims reports whether every condition matches the corresponding token
// claim. Conditions only match string-typed claims; a missing or non-string
// claim fails the policy.
func MatchClaims(conditions map[string]string, claims map[string]any) bool {
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

		if !MatchPattern(pattern, value) {
			return false
		}
	}

	return true
}
