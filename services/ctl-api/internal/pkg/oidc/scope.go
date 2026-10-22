package oidc

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

var ErrTooBroad = errors.New("the `sub` condition is too broad: it matches unrelated identities. " +
	"Put wildcards only after the part that identifies your repository or workload")

type Provider interface {
	IssuerURL() string
	Probes() []string
	ValidateConditions(v *validator.Validate, c Conditions) error
}

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

func TooBroad(pattern string, providerProbes []string) bool {
	for _, probe := range append(genericProbes(), providerProbes...) {
		if MatchPattern(pattern, probe) {
			return true
		}
	}

	return false
}
