package composite_error

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubError is a minimal CompositeError used only by pipeline tests.
type stubError struct {
	id string
}

func (s *stubError) Type() Type                      { return Type(s.id) }
func (s *stubError) Domain() Domain                  { return DomainNuon }
func (s *stubError) Severity() Severity              { return SeverityError }
func (s *stubError) Render(_ context.Context) Render { return Render{Title: s.id} }

// stubParser is a deterministic parser fixture.
type stubParser struct {
	name     string
	matches  bool
	emit     string
	contexts []ParseContext
	calls    int
	panics   bool
}

func (s *stubParser) Name() string             { return s.name }
func (s *stubParser) Version() string          { return "1" }
func (s *stubParser) Contexts() []ParseContext { return s.contexts }
func (s *stubParser) Parse(_ context.Context, _ ParseInput) ParseResult {
	s.calls++
	if s.panics {
		panic("boom")
	}
	if !s.matches {
		return ParseResult{Matched: false}
	}
	return ParseResult{Matched: true, Error: &stubError{id: s.emit}}
}

// staticLookup builds a ParserLookup from explicit specific/broad lists,
// emulating the catalog's most-specific-first iteration order.
func staticLookup(specific, broad []Parser) ParserLookup {
	return func(_ ParseContext) []Parser {
		out := append([]Parser{}, specific...)
		out = append(out, broad...)
		return out
	}
}

func unknownBuilder(_ ParseInput) ParseResult {
	return ParseResult{Matched: true, Error: &stubError{id: "unknown"}}
}

func TestPipeline_FirstMatchWinsAsPrimary(t *testing.T) {
	specific := &stubParser{name: "specific", matches: true, emit: "specific_err"}
	broad := &stubParser{name: "broad", matches: true, emit: "broad_err"}

	p := NewPipeline(staticLookup([]Parser{specific}, []Parser{broad}), unknownBuilder, nil)
	res := p.Parse(context.Background(), "terraform/plan", ParseInput{})

	require.NotNil(t, res.Primary.Error)
	assert.Equal(t, "specific_err", string(res.Primary.Error.Type()))
	require.Len(t, res.Secondaries, 1)
	assert.Equal(t, "broad_err", string(res.Secondaries[0].Error.Type()))
}

func TestPipeline_FallsBackToUnknownWhenNoMatch(t *testing.T) {
	a := &stubParser{name: "a", matches: false}
	b := &stubParser{name: "b", matches: false}

	p := NewPipeline(staticLookup([]Parser{a}, []Parser{b}), unknownBuilder, nil)
	res := p.Parse(context.Background(), "terraform", ParseInput{})

	require.NotNil(t, res.Primary.Error)
	assert.Equal(t, "unknown", string(res.Primary.Error.Type()))
	assert.Empty(t, res.Secondaries)
	assert.Equal(t, 1, a.calls)
	assert.Equal(t, 1, b.calls)
}

func TestPipeline_PanickingParserDoesNotBreakAndReportsToHandler(t *testing.T) {
	bomb := &stubParser{name: "bomb", panics: true}
	good := &stubParser{name: "good", matches: true, emit: "ok"}

	var caughtName string
	var caughtPanic any
	p := NewPipeline(staticLookup([]Parser{bomb}, []Parser{good}), unknownBuilder,
		func(name string, val any) { caughtName, caughtPanic = name, val })
	res := p.Parse(context.Background(), "terraform/apply", ParseInput{})

	require.NotNil(t, res.Primary.Error)
	assert.Equal(t, "ok", string(res.Primary.Error.Type()))
	assert.Equal(t, "bomb", caughtName)
	assert.Equal(t, "boom", caughtPanic)
}

func TestPipeline_SafetyNetWhenLookupNil(t *testing.T) {
	p := NewPipeline(nil, unknownBuilder, nil)
	res := p.Parse(context.Background(), "anything", ParseInput{})
	assert.Equal(t, "unknown", string(res.Primary.Error.Type()))
}

func TestPipeline_NonMatchedResultIgnored(t *testing.T) {
	half := &stubParser{name: "half", matches: false}
	good := &stubParser{name: "good", matches: true, emit: "got"}

	p := NewPipeline(staticLookup([]Parser{half, good}, nil), unknownBuilder, nil)
	res := p.Parse(context.Background(), "terraform", ParseInput{})

	assert.Equal(t, "got", string(res.Primary.Error.Type()))
	assert.Empty(t, res.Secondaries)
}
