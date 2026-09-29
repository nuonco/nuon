package errparse

import (
	"slices"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

type ParserFunc func(*ParseContext) compositeerrors.CompositeError

func NewParser(layer Layer, parse ParserFunc, opts ...Option) Parser {
	if parse == nil {
		panic("errparse: NewParser called with nil parse func")
	}
	p := &builtParser{layer: layer, parse: parse}
	for _, opt := range opts {
		if opt == nil {
			panic("errparse: NewParser called with a nil Option")
		}
		opt(p)
	}
	if p.alwaysCandidate == p.signalsSet {
		panic("errparse: parser must set exactly one of WithSignals or AlwaysCandidate")
	}
	return p
}

type Option func(*builtParser)

func WithTools(tools ...Tool) Option {
	cloned := slices.Clone(tools)
	return func(p *builtParser) { p.tools = cloned }
}

func WithSignals(signals ...string) Option {
	if len(signals) == 0 {
		panic("errparse: WithSignals called with no signals (use AlwaysCandidate for a signal-less parser)")
	}
	cloned := slices.Clone(signals)
	return func(p *builtParser) {
		p.signals = cloned
		p.signalsSet = true
	}
}

func AlwaysCandidate() Option {
	return func(p *builtParser) { p.alwaysCandidate = true }
}

func WithProviders(providers ...Provider) Option {
	if len(providers) == 0 {
		panic("errparse: WithProviders called with no providers")
	}
	for _, pr := range providers {
		if pr == ProviderUnknown {
			panic("errparse: WithProviders called with ProviderUnknown (an unresolved provider already fails open)")
		}
	}
	cloned := slices.Clone(providers)
	return func(p *builtParser) { p.providers = cloned }
}

type builtParser struct {
	layer     Layer
	tools     []Tool
	signals   []string
	providers []Provider
	parse     ParserFunc

	signalsSet      bool
	alwaysCandidate bool
}

var _ Parser = (*builtParser)(nil)

func (p *builtParser) Layer() Layer      { return p.layer }
func (p *builtParser) Tools() []Tool     { return p.tools }
func (p *builtParser) Signals() []string { return p.signals }

func (p *builtParser) Applicable(ctx *ParseContext) bool {
	if len(p.providers) == 0 {
		return true
	}
	provider := ctx.Provider()
	if provider == ProviderUnknown {
		return true
	}
	return slices.Contains(p.providers, provider)
}

func (p *builtParser) Parse(ctx *ParseContext) compositeerrors.CompositeError {
	return p.parse(ctx)
}
