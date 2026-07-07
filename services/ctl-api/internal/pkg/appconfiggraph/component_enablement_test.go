package appconfiggraph

import (
	"strconv"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func cccToggleable(id, name string, defaultEnabled bool, depIDs ...string) *app.ComponentConfigConnection {
	tr := true
	de := defaultEnabled
	return &app.ComponentConfigConnection{
		ComponentID:            id,
		Component:              app.Component{ID: id, Name: name},
		Toggleable:             &tr,
		DefaultEnabled:         &de,
		ComponentDependencyIDs: pq.StringArray(depIDs),
	}
}

func cccPlain(id, name string, depIDs ...string) *app.ComponentConfigConnection {
	return &app.ComponentConfigConnection{
		ComponentID:            id,
		Component:              app.Component{ID: id, Name: name},
		ComponentDependencyIDs: pq.StringArray(depIDs),
	}
}

// resolverFor returns a builder that takes per-component-name enabled toggles
// and materializes them as the reserved synthetic enabled inputs the resolver
// reads. In these tests component name == component id.
func resolverFor(cccs ...*app.ComponentConfigConnection) func(map[string]bool) *ComponentEnablementResolver {
	byID := make(map[string]*app.ComponentConfigConnection, len(cccs))
	for _, c := range cccs {
		byID[c.ComponentID] = c
	}
	return func(toggles map[string]bool) *ComponentEnablementResolver {
		inputs := make(map[string]*string, len(toggles))
		for name, enabled := range toggles {
			v := strconv.FormatBool(enabled)
			inputs[config.EnabledOverrideInputName(name)] = &v
		}
		return NewComponentEnablementResolver(byID, inputs)
	}
}

func TestEffectiveEnabled_OwnToggleAndDefault(t *testing.T) {
	r := resolverFor(
		cccToggleable("a", "a", true),
		cccToggleable("b", "b", false),
	)(nil)

	assert.True(t, r.EffectiveEnabled("a"), "default_enabled=true, no toggle -> enabled")
	assert.False(t, r.EffectiveEnabled("b"), "default_enabled=false, no toggle -> disabled")
}

func TestEffectiveEnabled_NonToggleableAlwaysOwnEnabled(t *testing.T) {
	r := resolverFor(cccPlain("a", "a"))(nil)
	assert.True(t, r.EffectiveEnabled("a"))
}

func TestEffectiveEnabled_CascadesThroughDeclaredDep(t *testing.T) {
	r := resolverFor(
		cccToggleable("a", "a", true),
		cccPlain("b", "b", "a"),
	)(map[string]bool{"a": false})

	assert.False(t, r.EffectiveEnabled("a"))
	assert.False(t, r.EffectiveEnabled("b"), "b depends on disabled a -> effectively disabled")
	assert.Equal(t, []string{"a"}, r.DisabledDependencies("b"))
}

func TestEffectiveEnabled_CascadesThroughOutputRef(t *testing.T) {
	b := cccPlain("b", "b")
	b.Refs = []refs.Ref{{Type: refs.RefTypeComponents, Name: "a", Value: "url"}}
	r := resolverFor(cccToggleable("a", "a", true), b)(map[string]bool{"a": false})

	assert.False(t, r.EffectiveEnabled("b"), "b output-refs disabled a -> effectively disabled")
}

func TestEffectiveEnabled_EnabledWhenDepEnabled(t *testing.T) {
	r := resolverFor(
		cccToggleable("a", "a", false),
		cccPlain("b", "b", "a"),
	)(map[string]bool{"a": true})

	assert.True(t, r.EffectiveEnabled("a"))
	assert.True(t, r.EffectiveEnabled("b"))
	assert.Empty(t, r.DisabledDependencies("b"))
}

func TestOwnEnabled_IgnoresDependencyClosure(t *testing.T) {
	r := resolverFor(
		cccToggleable("a", "a", true),
		cccPlain("b", "b", "a"),
	)(map[string]bool{"a": false})

	assert.False(t, r.OwnEnabled("a"))
	assert.True(t, r.OwnEnabled("b"), "b's own toggle is on even though its dep is disabled")
	assert.False(t, r.EffectiveEnabled("b"))
}

func TestTransitiveDependentsClosure_Chain(t *testing.T) {
	r := resolverFor(
		cccToggleable("a", "a", true),
		cccPlain("b", "b", "a"),
		cccPlain("c", "c", "b"),
	)(nil)

	assert.ElementsMatch(t, []string{"a", "b", "c"}, r.TransitiveDependentsClosure([]string{"a"}))
}

func TestDirectDependents(t *testing.T) {
	r := resolverFor(
		cccToggleable("a", "a", true),
		cccPlain("b", "b", "a"),
		cccPlain("c", "c", "a"),
		cccPlain("d", "d", "b"),
	)(nil)

	assert.Equal(t, []string{"b", "c"}, r.DirectDependents("a"))
	assert.Equal(t, []string{"d"}, r.DirectDependents("b"))
	assert.Empty(t, r.DirectDependents("d"))
}

func TestTopoSort_DepsFirstAndReverse(t *testing.T) {
	r := resolverFor(
		cccToggleable("a", "a", true),
		cccPlain("b", "b", "a"),
		cccPlain("c", "c", "b"),
	)(nil)

	assert.Equal(t, []string{"a", "b", "c"}, r.TopoSort([]string{"c", "a", "b"}))
	assert.Equal(t, []string{"c", "b", "a"}, r.ReverseTopoSort([]string{"c", "a", "b"}))
}

func TestEffectiveEnabled_CycleWithDisabledExternalDepFailsClosed(t *testing.T) {
	// a <-> b cycle, and a also depends on disabled c. The cycle must not let
	// b be cached as enabled; both should resolve effectively disabled.
	r := resolverFor(
		cccPlain("a", "a", "b", "c"),
		cccPlain("b", "b", "a"),
		cccToggleable("c", "c", true),
	)(map[string]bool{"c": false})

	assert.False(t, r.EffectiveEnabled("c"))
	assert.False(t, r.EffectiveEnabled("a"))
	assert.False(t, r.EffectiveEnabled("b"))
}

func TestEffectiveEnabled_Diamond(t *testing.T) {
	r := resolverFor(
		cccToggleable("a", "a", true),
		cccPlain("b", "b", "a"),
		cccPlain("c", "c", "a"),
		cccPlain("d", "d", "b", "c"),
	)(map[string]bool{"a": false})

	for _, id := range []string{"a", "b", "c", "d"} {
		assert.Falsef(t, r.EffectiveEnabled(id), "expected %s effectively disabled", id)
	}
	assert.ElementsMatch(t, []string{"a", "b", "c", "d"}, r.TransitiveDependentsClosure([]string{"a"}))
}
