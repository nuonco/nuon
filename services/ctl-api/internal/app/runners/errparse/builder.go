package errparse

import (
	"slices"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

// ParserFunc is the core recognise-and-classify behaviour of a parser: it
// attempts to produce a typed error from ctx, or returns nil to defer to the
// next candidate.
type ParserFunc func(*ParseContext) compositeerrors.CompositeError

// NewParser builds a Parser from its layer and parse behaviour plus optional
// facets. It covers the common case where a parser's gating is static, so a
// parser package registers a free function instead of a bespoke type with five
// one-line methods.
//
// The registration models a three-stage candidate pipeline (cheapest first),
// with layer as the tiebreak once several parsers match:
//
//	WithTools      structural bucket, considered before any text scan
//	WithSignals /  substrings that must be present (one cheap scan), or
//	AlwaysCandidate opt out of signal gating (the generic fallback)
//	WithProviders  lazy provider gate applied after a signal match
//	layer          precedence when several parsers still match
//
// layer and parse are required: there is no safe zero value for layer (the zero
// Layer is LayerProvider) and a nil parse panics. Exactly one of WithSignals or
// AlwaysCandidate must be given, so a parser that simply forgot its signals is
// rejected rather than silently running against every job.
//
// The Parser interface remains the underlying contract; implement it directly
// when a parser must hold state or gate on facets these options do not cover
// (Operation, Group, Owner).
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

// Option customizes a Parser built by NewParser.
type Option func(*builtParser)

// WithTools restricts the parser to the given execution tools, so it is never
// considered for a job run by another tool. With no WithTools the parser is
// tool-agnostic (considered for every job).
func WithTools(tools ...Tool) Option {
	cloned := slices.Clone(tools)
	return func(p *builtParser) { p.tools = cloned }
}

// WithSignals gates the parser on substrings that must be present in the raw
// text; the parser's Parse only runs once one is found. At least one signal is
// required (use AlwaysCandidate for a deliberately signal-less parser).
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

// AlwaysCandidate opts the parser out of signal gating so it is a candidate for
// every job. It exists for the generic fallback; a normal parser should gate on
// WithSignals so it is not run against unrelated failures.
func AlwaysCandidate() Option {
	return func(p *builtParser) { p.alwaysCandidate = true }
}

// WithProviders gates the parser on the cloud provider the job targeted. The
// gate is applied after a signal match (provider resolution is a lazy lookup)
// and fails open: a job whose provider cannot be resolved still passes, so
// provider gating is only ever a false-positive guard. Passing no providers or
// ProviderUnknown is a configuration mistake and panics.
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

// builtParser is the Parser implementation backing NewParser.
type builtParser struct {
	layer     Layer
	tools     []Tool
	signals   []string
	providers []Provider
	parse     ParserFunc

	// signalsSet and alwaysCandidate record which signal-gating option was
	// chosen so NewParser can require exactly one. An empty signals slice is
	// otherwise ambiguous between "forgot signals" and "deliberately none".
	signalsSet      bool
	alwaysCandidate bool
}

var _ Parser = (*builtParser)(nil)

func (p *builtParser) Layer() Layer      { return p.layer }
func (p *builtParser) Tools() []Tool     { return p.tools }
func (p *builtParser) Signals() []string { return p.signals }

// Applicable applies the declared provider gate, failing open on an
// unresolved provider so the fail-open contract lives in one place instead of
// being hand-rolled by every parser.
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
