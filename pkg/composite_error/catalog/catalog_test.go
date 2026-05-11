package catalog

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	composite_error "github.com/nuonco/nuon/pkg/composite_error"
)

// fakeError is a minimal CompositeError for catalog tests.
type fakeError struct {
	Msg string `json:"msg"`
}

func (f *fakeError) Type() composite_error.Type         { return "test_fake" }
func (f *fakeError) Domain() composite_error.Domain     { return composite_error.DomainNuon }
func (f *fakeError) Severity() composite_error.Severity { return composite_error.SeverityError }
func (f *fakeError) Render(_ context.Context) composite_error.Render {
	return composite_error.Render{Title: f.Msg}
}

type fakeParser struct {
	name     string
	contexts []composite_error.ParseContext
}

func (p fakeParser) Name() string                             { return p.name }
func (p fakeParser) Version() string                          { return "1" }
func (p fakeParser) Contexts() []composite_error.ParseContext { return p.contexts }
func (p fakeParser) Parse(_ context.Context, _ composite_error.ParseInput) composite_error.ParseResult {
	return composite_error.ParseResult{Matched: false}
}

func TestRegisterTypeAndHydrate(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	RegisterType("test_fake", func() composite_error.CompositeError { return &fakeError{} })

	got, err := Hydrate("test_fake", []byte(`{"msg":"hello"}`))
	require.NoError(t, err)
	require.NotNil(t, got)

	fe, ok := got.(*fakeError)
	require.True(t, ok)
	assert.Equal(t, "hello", fe.Msg)
}

func TestRegisterTypePanicsOnDuplicate(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	RegisterType("test_fake", func() composite_error.CompositeError { return &fakeError{} })
	assert.Panics(t, func() {
		RegisterType("test_fake", func() composite_error.CompositeError { return &fakeError{} })
	})
}

func TestHydrateUnknownTypeErrors(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	_, err := Hydrate("not_registered", []byte(`{}`))
	assert.Error(t, err)
}

func TestHydrateEmptyDataReturnsZeroValue(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	RegisterType("test_fake", func() composite_error.CompositeError { return &fakeError{} })

	got, err := Hydrate("test_fake", nil)
	require.NoError(t, err)
	require.NotNil(t, got)

	fe := got.(*fakeError)
	assert.Empty(t, fe.Msg)
}

func TestRegisterParserPanicsWithoutContexts(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	assert.Panics(t, func() {
		RegisterParser(fakeParser{name: "no-ctx"})
	})
}

func TestParsersForContextWalksAncestors(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	root := fakeParser{name: "root", contexts: []composite_error.ParseContext{""}}
	tf := fakeParser{name: "tf", contexts: []composite_error.ParseContext{"terraform"}}
	tfPlan := fakeParser{name: "tf-plan", contexts: []composite_error.ParseContext{"terraform/plan"}}
	helm := fakeParser{name: "helm", contexts: []composite_error.ParseContext{"helm"}}

	RegisterParser(root)
	RegisterParser(tf)
	RegisterParser(tfPlan)
	RegisterParser(helm)

	got := ParsersForContext("terraform/plan")
	require.Len(t, got, 3, "should return tf-plan + tf + root, not helm")
	assert.Equal(t, "tf-plan", got[0].Name())
	assert.Equal(t, "tf", got[1].Name())
	assert.Equal(t, "root", got[2].Name())
}

func TestParserCanRegisterAcrossSubtrees(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	cross := fakeParser{
		name:     "aws-perm",
		contexts: []composite_error.ParseContext{"terraform", "helm", "runner/job"},
	}
	RegisterParser(cross)

	for _, ctx := range []composite_error.ParseContext{
		"terraform/plan", "terraform/apply", "helm/install", "runner/job",
	} {
		got := ParsersForContext(ctx)
		require.Len(t, got, 1, ctx)
		assert.Equal(t, "aws-perm", got[0].Name())
	}
	assert.Empty(t, ParsersForContext("kubernetes/rollout"))
}

func TestAncestorsOf(t *testing.T) {
	cases := []struct {
		in   composite_error.ParseContext
		want []composite_error.ParseContext
	}{
		{"", []composite_error.ParseContext{""}},
		{"terraform", []composite_error.ParseContext{"terraform", ""}},
		{"terraform/plan", []composite_error.ParseContext{"terraform/plan", "terraform", ""}},
		{"a/b/c", []composite_error.ParseContext{"a/b/c", "a/b", "a", ""}},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, ancestorsOf(tc.in), string(tc.in))
	}
}

func TestTypesReturnsSortedRegisteredTypes(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	RegisterType("zeta", func() composite_error.CompositeError { return &fakeError{} })
	RegisterType("alpha", func() composite_error.CompositeError { return &fakeError{} })
	RegisterType("mu", func() composite_error.CompositeError { return &fakeError{} })

	got := Types()
	assert.Equal(t, []composite_error.Type{"alpha", "mu", "zeta"}, got)
}
