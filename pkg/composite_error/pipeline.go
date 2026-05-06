package composite_error

import (
	"context"
	"fmt"
)

// ParserLookup returns parsers applicable to a ParseContext, ordered
// most-specific → least-specific. The catalog package implements this against
// its in-memory registry; tests can substitute a stub.
type ParserLookup func(ParseContext) []Parser

// UnknownErrorBuilder constructs the safety-net result when no parser matches.
// Provided by the unknown_error type's package via SetUnknownBuilder().
type UnknownErrorBuilder func(in ParseInput) ParseResult

// Pipeline dispatches ParseInput to registered parsers and produces a
// PipelineResult: a primary CompositeError plus zero-or-more secondaries.
//
// Pipeline is stateless and safe for concurrent use.
type Pipeline struct {
	lookup  ParserLookup
	unknown UnknownErrorBuilder
}

// NewPipeline constructs a Pipeline. Both arguments are required:
//
//   - lookup is typically catalog.ParsersForContext.
//   - unknown is typically the factory the unknown_error package exposes.
func NewPipeline(lookup ParserLookup, unknown UnknownErrorBuilder) *Pipeline {
	return &Pipeline{lookup: lookup, unknown: unknown}
}

// PipelineResult is the output of Pipeline.Parse.
type PipelineResult struct {
	// Primary is the headline error to attach to the owner. Always non-nil.
	Primary ParseResult

	// Secondaries are additional matches at any level that the caller may
	// also persist as separate rows on the same owner.
	Secondaries []ParseResult
}

// Parse runs the parsers registered for ctx against in and returns a
// PipelineResult.
//
// Dispatch rule:
//
//  1. Walk ancestors of ctx, most-specific → least-specific.
//  2. At each level, run every registered parser in registration order.
//  3. The first matching result at the most-specific level becomes Primary.
//  4. All other matches at any level become Secondaries.
//  5. If nothing matches, fall back to the unknown error builder.
//
// Parser panics are recovered and treated as a non-match.
func (p *Pipeline) Parse(ctx context.Context, parseCtx ParseContext, in ParseInput) PipelineResult {
	if p.lookup == nil {
		return PipelineResult{Primary: p.fallback(in)}
	}

	parsers := p.lookup(parseCtx)

	var primary *ParseResult
	var secondaries []ParseResult

	for _, parser := range parsers {
		res := safeParse(ctx, parser, in)
		if !res.Matched || res.Error == nil {
			continue
		}
		if primary == nil {
			primary = &res
			continue
		}
		secondaries = append(secondaries, res)
	}

	if primary == nil {
		return PipelineResult{Primary: p.fallback(in)}
	}

	return PipelineResult{Primary: *primary, Secondaries: secondaries}
}

// fallback returns the unknown_error result. If no builder is wired, returns
// a Matched=false envelope so callers can decide what to do (helpers will
// upgrade this into a synthetic unknown row).
func (p *Pipeline) fallback(in ParseInput) ParseResult {
	if p.unknown != nil {
		return p.unknown(in)
	}
	return ParseResult{Matched: false}
}

func safeParse(ctx context.Context, parser Parser, in ParseInput) (out ParseResult) {
	defer func() {
		if r := recover(); r != nil {
			// A parser panic must never break the failure path.
			// Log via the standard library so we don't bring a logger
			// into the core package; helpers add structured logging.
			fmt.Printf("composite_error: parser %q panicked: %v\n", parser.Name(), r)
			out = ParseResult{Matched: false}
		}
	}()
	return parser.Parse(ctx, in)
}
