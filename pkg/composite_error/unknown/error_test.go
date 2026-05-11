package unknown

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	composite_error "github.com/nuonco/nuon/pkg/composite_error"
	"github.com/nuonco/nuon/pkg/composite_error/catalog"
)

func TestBuild_PopulatesMessageFromFirstLine(t *testing.T) {
	res := Build(composite_error.ParseInput{Raw: []byte("first line\nsecond line\n")})
	require.True(t, res.Matched)

	e := res.Error.(*Error)
	assert.Equal(t, "first line", e.Message)
	assert.Equal(t, "first line\nsecond line", e.Details)
}

func TestBuild_FallsBackToGoError(t *testing.T) {
	res := Build(composite_error.ParseInput{GoErr: assertErr("hi")})
	e := res.Error.(*Error)
	assert.Equal(t, "hi", e.Message)
	assert.Equal(t, "hi", e.Details)
}

func TestBuild_CapturesExitCode(t *testing.T) {
	res := Build(composite_error.ParseInput{Raw: []byte("oh no"), ExitCode: 42})
	e := res.Error.(*Error)
	require.NotNil(t, e.ExitCode)
	assert.Equal(t, 42, *e.ExitCode)
	require.NotNil(t, res.Source.ExitCode)
	assert.Equal(t, 42, *res.Source.ExitCode)
}

func TestBuild_PopulatesSourceMetadata(t *testing.T) {
	res := Build(composite_error.ParseInput{
		Raw:   []byte("err"),
		GoErr: assertErr("wrapper"),
	})
	assert.Equal(t, "unknown_error.builder", res.Source.ParserName)
	assert.Equal(t, "1", res.Source.ParserVersion)
	assert.Equal(t, "wrapper", res.Source.GoError)
	assert.Equal(t, "err", res.Source.Snippet)
}

func TestRender_DefaultTitle(t *testing.T) {
	r := (&Error{}).Render(context.Background())
	assert.Equal(t, "An unknown error occurred", r.Title)
}

func TestRender_DetailsSectionWhenMultiLine(t *testing.T) {
	e := &Error{Message: "first line", Details: "first line\nsecond line"}
	r := e.Render(context.Background())
	assert.Equal(t, "first line", r.Title)
	require.Len(t, r.Sections, 1)
	assert.Equal(t, "Error details", r.Sections[0].Heading)
	assert.Equal(t, "first line\nsecond line", r.Sections[0].Body)
}

func TestRender_NoDetailsSectionWhenIdenticalToMessage(t *testing.T) {
	e := &Error{Message: "boom", Details: "boom"}
	r := e.Render(context.Background())
	assert.Empty(t, r.Sections)
}

// TestRoundTrip_EncodeStoreHydrateRender exercises the contract the DB layer
// will rely on: marshal a typed error → store as JSON → hydrate via the
// catalog → render. This is the spec's Phase 1 round-trip test.
func TestRoundTrip_EncodeStoreHydrateRender(t *testing.T) {
	original := &Error{Message: "boom"}
	exit := 7
	original.ExitCode = &exit

	stored, err := json.Marshal(original)
	require.NoError(t, err)

	hydrated, err := catalog.Hydrate(Type, stored)
	require.NoError(t, err)
	require.NotNil(t, hydrated)

	rendered := hydrated.Render(context.Background())
	assert.Equal(t, "boom", rendered.Title)
	require.Len(t, rendered.Sections, 1)
	assert.Equal(t, "Exit code", rendered.Sections[0].Heading)
	assert.Equal(t, "7", rendered.Sections[0].Body)
}

type assertErr string

func (a assertErr) Error() string { return string(a) }
