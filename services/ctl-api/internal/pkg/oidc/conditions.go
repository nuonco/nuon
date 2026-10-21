package oidc

type Conditions struct {
	Sub   string            `validate:"required,max=512,oidc_pattern"`
	Extra map[string]string `validate:"dive,keys,oidc_claim_name,endkeys,required,max=512,oidc_pattern"`
}

func ParseConditions(raw map[string]string) Conditions {
	c := Conditions{Sub: raw["sub"]}

	for name, pattern := range raw {
		if name == "sub" {
			continue
		}
		if c.Extra == nil {
			c.Extra = make(map[string]string, len(raw))
		}
		c.Extra[name] = pattern
	}

	return c
}

// Match reports whether every condition matches the corresponding token
// claim. A missing or non-string claim fails, and empty conditions never
// match.
func (c Conditions) Match(claims Claims) bool {
	if c.Sub == "" && len(c.Extra) == 0 {
		return false
	}

	if c.Sub != "" && !matchClaim(claims, "sub", c.Sub) {
		return false
	}

	for name, pattern := range c.Extra {
		if !matchClaim(claims, name, pattern) {
			return false
		}
	}

	return true
}

func matchClaim(claims Claims, name, pattern string) bool {
	value, ok := claims.Lookup(name)
	if !ok {
		return false
	}

	return MatchPattern(pattern, value)
}
