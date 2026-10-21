package oidc

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ErrTooBroad is returned when a `sub` pattern matches unrelated identities.
var ErrTooBroad = errors.New("the `sub` condition is too broad: it matches unrelated identities. " +
	"Put wildcards only after the part that identifies your repository or workload")

// Provider holds the rules specific to one identity provider.
type Provider interface {
	// IssuerURL is the issuer the provider's rules apply to.
	IssuerURL() string
	// Probes returns provider-shaped `sub` values of identities no policy
	// should match, extending the generic probes.
	Probes() []string
	// ValidateConditions applies provider rules on top of the generic ones.
	ValidateConditions(v *validator.Validate, c Conditions) error
}

// Providers looks up the Provider for an issuer.
type Providers struct {
	byIssuer map[string]Provider
}

func NewProviders(providers ...Provider) Providers {
	p := Providers{byIssuer: make(map[string]Provider, len(providers))}
	for _, provider := range providers {
		p.byIssuer[normalizeIssuer(provider.IssuerURL())] = provider
	}

	return p
}

// For returns the provider for issuerURL, or the generic provider, which adds
// no rules, when none is registered.
func (p Providers) For(issuerURL string) Provider {
	if provider, ok := p.byIssuer[normalizeIssuer(issuerURL)]; ok {
		return provider
	}

	return genericProvider{issuerURL: issuerURL}
}

func normalizeIssuer(issuerURL string) string {
	return strings.TrimSuffix(issuerURL, "/")
}

type genericProvider struct {
	issuerURL string
}

func (g genericProvider) IssuerURL() string { return g.issuerURL }

func (genericProvider) Probes() []string { return nil }

func (genericProvider) ValidateConditions(*validator.Validate, Conditions) error { return nil }

// Probe values stand in for identities a policy author has never heard of.
// They are fixed rather than random so validation is deterministic.
const (
	ProbeA = "nuonprobeqz7a"
	ProbeB = "nuonprobekx3b"
	ProbeC = "nuonprobewv5c"
	ProbeD = "nuonprobejm9d"
)

func genericProbes() []string {
	return []string{
		ProbeA,
		ProbeA + ":" + ProbeB,
		ProbeA + ":" + ProbeB + ":" + ProbeC,
		ProbeA + ":" + ProbeB + ":" + ProbeC + ":" + ProbeD,
		ProbeA + "/" + ProbeB,
	}
}

// TooBroad reports whether the `sub` pattern matches foreign probe values,
// which no policy for a specific identity should.
func TooBroad(pattern string, providerProbes []string) bool {
	for _, probe := range append(genericProbes(), providerProbes...) {
		if MatchPattern(pattern, probe) {
			return true
		}
	}

	return false
}
