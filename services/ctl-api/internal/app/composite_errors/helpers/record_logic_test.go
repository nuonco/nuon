package helpers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	composite_error "github.com/nuonco/nuon/pkg/composite_error"
)

// stubError is a minimal CompositeError used by the record-logic tests.
// We don't need a registered catalog entry here — the logic under test
// (recordInputFromParseResult) operates on the typed instance directly.
type stubError struct{ id string }

func (s *stubError) Type() composite_error.Type         { return composite_error.Type(s.id) }
func (s *stubError) Domain() composite_error.Domain     { return composite_error.DomainNuon }
func (s *stubError) Severity() composite_error.Severity { return composite_error.SeverityError }
func (s *stubError) Render(_ context.Context) composite_error.Render {
	return composite_error.Render{Title: s.id}
}

func TestRecordInputFromParseResult_FlattensTreeWithPrimaryFirstCause(t *testing.T) {
	res := composite_error.PipelineResult{
		Primary: composite_error.ParseResult{
			Matched: true,
			Error:   &stubError{id: "root"},
			Causes: []composite_error.ParseResult{
				{Matched: true, Error: &stubError{id: "child1"}},
				{Matched: true, Error: &stubError{id: "child2"}},
			},
		},
		Secondaries: []composite_error.ParseResult{
			{Matched: true, Error: &stubError{id: "secondary"}},
		},
	}

	in := recordInputFromParseResult("workflow_steps", "iws_xyz", res)

	assert.Equal(t, "workflow_steps", in.OwnerType)
	assert.Equal(t, "iws_xyz", in.OwnerID)
	assert.Equal(t, "root", string(in.Error.Type()))

	require.Len(t, in.Causes, 3, "primary children + secondaries become root causes")
	assert.Equal(t, "child1", string(in.Causes[0].Error.Type()))
	assert.True(t, in.Causes[0].IsPrimary, "first child of primary is the primary cause")

	assert.Equal(t, "child2", string(in.Causes[1].Error.Type()))
	assert.False(t, in.Causes[1].IsPrimary)

	assert.Equal(t, "secondary", string(in.Causes[2].Error.Type()))
	assert.False(t, in.Causes[2].IsPrimary, "cross-parser secondaries never get IsPrimary")
}

func TestRecordCauseFromParseResult_RecursesAndOnlyFirstIsPrimary(t *testing.T) {
	r := composite_error.ParseResult{
		Matched: true,
		Error:   &stubError{id: "lvl0"},
		Causes: []composite_error.ParseResult{
			{Matched: true, Error: &stubError{id: "lvl1a"}},
			{Matched: true, Error: &stubError{id: "lvl1b"}, Causes: []composite_error.ParseResult{
				{Matched: true, Error: &stubError{id: "lvl2"}},
			}},
		},
	}

	c := recordCauseFromParseResult(r, true)

	assert.Equal(t, "lvl0", string(c.Error.Type()))
	assert.True(t, c.IsPrimary)

	require.Len(t, c.Causes, 2)
	assert.Equal(t, "lvl1a", string(c.Causes[0].Error.Type()))
	assert.True(t, c.Causes[0].IsPrimary, "first cause of a parent is primary")
	assert.False(t, c.Causes[1].IsPrimary)

	require.Len(t, c.Causes[1].Causes, 1)
	assert.True(t, c.Causes[1].Causes[0].IsPrimary, "primary recurses depth-wise on first child")
}

func TestValidateRecordInput_RequiresOwnerAndError(t *testing.T) {
	cases := []struct {
		name string
		in   RecordInput
	}{
		{"missing owner type", RecordInput{OwnerID: "x", Error: &stubError{id: "e"}}},
		{"missing owner id", RecordInput{OwnerType: "x", Error: &stubError{id: "e"}}},
		{"missing error", RecordInput{OwnerType: "x", OwnerID: "y"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Error(t, validateRecordInput(tc.in))
		})
	}
}
