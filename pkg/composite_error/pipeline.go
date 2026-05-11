package composite_error

import "context"

// ParserLookup returns parsers applicable to a ParseContext, ordered
// most-specific → least-specific. The catalog package implements this against
// its in-memory registry; tests can substitute a stub.
type ParserLookup func(ParseContext) []Parser

// UnknownErrorBuilder constructs the safety-net result when no parser matches.
// Provided by the unknown_error builtin via NewPipeline.
type UnknownErrorBuilder func(in ParseInput) ParseResult

// PanicHandler is invoked when a parser panics. The pipeline always treats
// the panicking parser's call as a non-match; the handler is for observability.
// nil is a no-op.
type PanicHandler func(parserName string, panicValue any)

// Pipeline dispatches ParseInput to registered parsers and produces a
// PipelineResult: a primary CompositeError plus zero-or-more secondaries.
//
// Pipeline is stateless and safe for concurrent use.
type Pipeline struct {
	lookup  ParserLookup
	unknown UnknownErrorBuilder
	onPanic PanicHandler
}

// NewPipeline constructs a Pipeline. Both lookup and unknown are required.
//
//   - lookup is typically catalog.ParsersForContext.
//   - unknown is typically unknown_error.Build.
func NewPipeline(lookup ParserLookup, unknown UnknownErrorBuilder, onPanic PanicHandler) *Pipeline {
	return &Pipeline{lookup: lookup, unknown: unknown, onPanic: onPanic}
}

// PipelineResult is the output of Pipeline.Parse.
type PipelineResult struct {
	// Primary is the headline error to attach to the owner. Always non-nil
	// after Parse() (the pipeline guarantees a fallback).
	Primary ParseResult

	// Secondaries are additional matches at any level that the caller may
	// also persist as separate rows on the same owner.
	Secondaries []ParseResult
}

// Parse runs the parsers registered for parseCtx against in and returns a
// PipelineResult.
//
// Dispatch rule:
//
//  1. Walk ancestors of parseCtx, most-specific → least-specific (via the
//     ParserLookup).
//  2. At each level, run every registered parser in registration order.
//  3. The first matching result becomes Primary.
//  4. All other matches at any level become Secondaries.
//  5. If nothing matches, fall back to the unknown error builder.
//
// Parser panics are recovered, treated as a non-match, and reported via
// the onPanic handler if one was supplied.
func (p *Pipeline) Parse(ctx context.Context, parseCtx ParseContext, in ParseInput) PipelineResult {
	if p.lookup == nil {
		return PipelineResult{Primary: p.fallback(in)}
	}

	parsers := p.lookup(parseCtx)

	var primary *ParseResult
	var secondaries []ParseResult

	for _, parser := range parsers {
		res := p.safeParse(ctx, parser, in)
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

// fallback returns the unknown_error result.
func (p *Pipeline) fallback(in ParseInput) ParseResult {
	if p.unknown != nil {
		return p.unknown(in)
	}
	return ParseResult{Matched: false}
}

func (p *Pipeline) safeParse(ctx context.Context, parser Parser, in ParseInput) (out ParseResult) {
	defer func() {
		if r := recover(); r != nil {
			if p.onPanic != nil {
				p.onPanic(parser.Name(), r)
			}
			out = ParseResult{Matched: false}
		}
	}()
	return parser.Parse(ctx, in)
}
