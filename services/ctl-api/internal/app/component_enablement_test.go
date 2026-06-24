package app

import (
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/refs"
)

func cccToggleable(id, name string, defaultEnabled bool, depIDs ...string) *ComponentConfigConnection {
	tr := true
	de := defaultEnabled
	return &ComponentConfigConnection{
		ComponentID:            id,
		Component:              Component{ID: id, Name: name},
		Toggleable:             &tr,
		DefaultEnabled:         &de,
		ComponentDependencyIDs: pq.StringArray(depIDs),
	}
}

func cccPlain(id, name string, depIDs ...string) *ComponentConfigConnection {
	return &ComponentConfigConnection{
		ComponentID:            id,
		Component:              Component{ID: id, Name: name},
		ComponentDependencyIDs: pq.StringArray(depIDs),
	}
}

func resolverFor(cccs ...*ComponentConfigConnection) func(map[string]*string) *ComponentEnablementResolver {
	byID := make(map[string]*ComponentConfigConnection, len(cccs))
	for _, c := range cccs {
		byID[c.ComponentID] = c
	}
	return func(inputs map[string]*string) *ComponentEnablementResolver {
		return NewComponentEnablementResolver(byID, inputs)
	}
}

func enabledInput(name string, enabled bool) (string, *string) {
	v := "false"
	if enabled {
		v = "true"
	}
	return config.EnabledOverrideInputName(name), &v
}

func TestEffectiveEnabled_OwnToggleAndDefault(t *testing.T) {
	r := resolverFor(
		cccToggleable("a", "a", true),
		cccToggleable("b", "b", false),
	)(nil)

	assert.True(t, r.EffectiveEnabled("a"), "default_enabled=true, no input -> enabled")
	assert.False(t, r.EffectiveEnabled("b"), "default_enabled=false, no input -> disabled")
}

func TestEffectiveEnabled_NonToggleableAlwaysOwnEnabled(t *testing.T) {
	r := resolverFor(cccPlain("a", "a"))(nil)
	assert.True(t, r.EffectiveEnabled("a"))
}

func TestEffectiveEnabled_CascadesThroughDeclaredDep(t *testing.T) {
	k, v := enabledInput("a", false)
	r := resolverFor(
		cccToggleable("a", "a", true),
		cccPlain("b", "b", "a"),
	)(map[string]*string{k: v})

	assert.False(t, r.EffectiveEnabled("a"))
	assert.False(t, r.EffectiveEnabled("b"), "b depends on disabled a -> effectively disabled")
	assert.Equal(t, []string{"a"}, r.DisabledDependencies("b"))
}

func TestEffectiveEnabled_CascadesThroughOutputRef(t *testing.T) {
	k, v := enabledInput("a", false)
	b := cccPlain("b", "b")
	b.Refs = []refs.Ref{{Type: refs.RefTypeComponents, Name: "a", Value: "url"}}
	r := resolverFor(cccToggleable("a", "a", true), b)(map[string]*string{k: v})

	assert.False(t, r.EffectiveEnabled("b"), "b output-refs disabled a -> effectively disabled")
}

func TestEffectiveEnabled_EnabledWhenDepEnabled(t *testing.T) {
	k, v := enabledInput("a", true)
	r := resolverFor(
		cccToggleable("a", "a", false),
		cccPlain("b", "b", "a"),
	)(map[string]*string{k: v})

	assert.True(t, r.EffectiveEnabled("a"))
	assert.True(t, r.EffectiveEnabled("b"))
	assert.Empty(t, r.DisabledDependencies("b"))
}

func TestTransitiveDependentsClosure_Chain(t *testing.T) {
	r := resolverFor(
		cccToggleable("a", "a", true),
		cccPlain("b", "b", "a"),
		cccPlain("c", "c", "b"),
	)(nil)

	assert.ElementsMatch(t, []string{"a", "b", "c"}, r.TransitiveDependentsClosure([]string{"a"}))
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
	k, v := enabledInput("c", false)
	r := resolverFor(
		cccPlain("a", "a", "b", "c"),
		cccPlain("b", "b", "a"),
		cccToggleable("c", "c", true),
	)(map[string]*string{k: v})

	assert.False(t, r.EffectiveEnabled("c"))
	assert.False(t, r.EffectiveEnabled("a"))
	assert.False(t, r.EffectiveEnabled("b"))
}

func TestEffectiveEnabled_Diamond(t *testing.T) {
	k, v := enabledInput("a", false)
	r := resolverFor(
		cccToggleable("a", "a", true),
		cccPlain("b", "b", "a"),
		cccPlain("c", "c", "a"),
		cccPlain("d", "d", "b", "c"),
	)(map[string]*string{k: v})

	for _, id := range []string{"a", "b", "c", "d"} {
		assert.Falsef(t, r.EffectiveEnabled(id), "expected %s effectively disabled", id)
	}
	assert.ElementsMatch(t, []string{"a", "b", "c", "d"}, r.TransitiveDependentsClosure([]string{"a"}))
}
