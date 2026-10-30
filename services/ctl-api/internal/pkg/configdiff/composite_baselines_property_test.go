package configdiff

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	"hegel.dev/go/hegel"

	"github.com/nuonco/nuon/pkg/config"
	pkgdiff "github.com/nuonco/nuon/pkg/config/diff"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type appliedComponentFact struct {
	present bool
	kind    config.ComponentType
	value   string
}

type appliedConfigFact struct {
	components  []appliedComponentFact
	sandbox     string
	stack       string
	runner      string
	description string
}

func drawAppliedConfigFact(t *hegel.T, n int) appliedConfigFact {
	f := appliedConfigFact{
		sandbox:     hegel.Draw(t, hegel.SampledFrom([]string{"1.6.0", "1.7.0"})),
		stack:       hegel.Draw(t, hegel.SampledFrom([]string{"primary", "secondary"})),
		runner:      hegel.Draw(t, hegel.SampledFrom([]string{"ecs", "eks"})),
		description: hegel.Draw(t, hegel.SampledFrom([]string{"old", "new"})),
	}
	for range n {
		f.components = append(f.components, appliedComponentFact{
			present: hegel.Draw(t, hegel.Booleans()),
			kind:    hegel.Draw(t, hegel.SampledFrom([]config.ComponentType{config.JobComponentType, config.KubernetesManifestComponentType})),
			value:   hegel.Draw(t, hegel.SampledFrom([]string{"alpha", "beta", "gamma"})),
		})
	}
	return f
}

func drawShuffled[T any](t *hegel.T, values []T) []T {
	values = slices.Clone(values)
	for i := len(values) - 1; i > 0; i-- {
		j := hegel.Draw(t, hegel.Integers(0, i))
		values[i], values[j] = values[j], values[i]
	}
	return values
}

func (f appliedConfigFact) component(i int) appliedComponentFact {
	if i >= len(f.components) || !f.components[i].present {
		return appliedComponentFact{}
	}
	return f.components[i]
}

func (f appliedConfigFact) materialize() *config.AppConfig {
	cfg := &config.AppConfig{
		Description: f.description,
		Sandbox:     &config.AppSandboxConfig{TerraformVersion: f.sandbox},
		Stack:       &config.StackConfig{Name: f.stack, Type: "aws-cloudformation"},
		Runner:      &config.AppRunnerConfig{RunnerType: f.runner},
	}
	for i := range f.components {
		c := f.component(i)
		if !c.present {
			continue
		}
		component := &config.Component{Name: fmt.Sprintf("resource-%d", i), Type: c.kind, VarName: c.value}
		if c.kind == config.JobComponentType {
			component.Job = &config.JobComponentConfig{ImageURL: "example.com/acme/job", Tag: "v1"}
		} else {
			component.KubernetesManifest = &config.KubernetesManifestComponentConfig{Namespace: "acme", Manifest: "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: acme"}
		}
		cfg.Components = append(cfg.Components, component)
	}
	return cfg
}

func (f appliedConfigFact) persisted() *app.AppConfig {
	n := 0
	for _, c := range f.components {
		if c.present {
			n++
		}
	}
	cfg := &app.AppConfig{}
	if n > 0 {
		cfg.ComponentConfigConnections = make([]app.ComponentConfigConnection, 0, n)
	}
	for i := range f.components {
		c := f.component(i)
		if c.present {
			cfg.ComponentConfigConnections = append(cfg.ComponentConfigConnections, app.ComponentConfigConnection{
				ComponentID: fmt.Sprintf("cmp-%d", i), ComponentName: fmt.Sprintf("resource-%d", i),
				Type: app.ComponentType(c.kind), Checksum: fmt.Sprintf("%s/%s", c.kind, c.value),
			})
		}
	}
	return cfg
}

func compositeFactInputs(facts map[string]appliedConfigFact, baselines CompositeBaselines, n int) compositeTreeInputs {
	in := compositeTreeInputs{
		New: facts["desired"].materialize(), Baselines: baselines,
		Intermediates: map[string]*config.AppConfig{}, ComponentIDsByName: map[string]string{},
	}
	for id, fact := range facts {
		in.Intermediates[id] = fact.materialize()
	}
	for i := range n {
		in.ComponentIDsByName[fmt.Sprintf("resource-%d", i)] = fmt.Sprintf("cmp-%d", i)
	}
	return in
}

func requireStringFact(t require.TestingT, node *pkgdiff.Diff, key, old, next string) {
	field := topChild(node, key)
	require.NotNil(t, field, key)
	require.NotNil(t, field.Diff, key)
	op := pkgdiff.OpChange
	display := fmt.Sprintf("'%s' -> '%s'", old, next)
	switch {
	case old == next:
		op, display = pkgdiff.OpNoop, fmt.Sprintf("'%s' (unchanged)", old)
	case old == "":
		op = pkgdiff.OpAdd
	case next == "":
		op = pkgdiff.OpRemove
	}
	require.Equal(t, op, field.Diff.Op, "%s: %q -> %q", key, old, next)
	require.Equal(t, display, field.Diff.Diff, key)
}

func authoredObservation(tree *pkgdiff.Diff) *pkgdiff.Diff {
	if tree == nil {
		return nil
	}
	out := &pkgdiff.Diff{Key: tree.Key, ResourceID: tree.ResourceID, Diff: tree.Diff}
	for _, child := range tree.Children {
		out.Children = append(out.Children, authoredObservation(child))
	}
	return out
}

func requireCompositeFacts(t *hegel.T, facts map[string]appliedConfigFact, baselines CompositeBaselines, n int) *pkgdiff.Diff {
	in := compositeFactInputs(facts, baselines, n)
	before, err := json.Marshal(in)
	require.NoError(t, err)
	tree := spliceCompositeAppConfigTree(in)
	desired := facts["desired"]
	requireStringFact(t, tree, "description", facts[baselines.Fallback].description, desired.description)
	requireStringFact(t, topChild(tree, "sandbox"), "terraform_version", facts[baselines.Sandbox].sandbox, desired.sandbox)
	requireStringFact(t, topChild(tree, "stack"), "name", facts[baselines.Stack].stack, desired.stack)
	requireStringFact(t, topChild(tree, "runner"), "runner_type", facts[baselines.Stack].runner, desired.runner)
	loaded := map[string]*app.AppConfig{}
	for id, fact := range facts {
		loaded[id] = fact.persisted()
	}
	summary := &app.InstallConfigDiff{}
	require.NoError(t, fillComponentSlices(summary, baselines, loaded, desired.persisted(), func(string) (*pkgdiff.Diff, error) { return nil, nil }))
	wantCategories := map[string][]string{"added": {}, "removed": {}, "changed": {}, "unchanged": {}}
	for i := range n {
		id, name := fmt.Sprintf("cmp-%d", i), fmt.Sprintf("resource-%d", i)
		baseline, ok := baselines.Components[id]
		if !ok {
			baseline = baselines.Fallback
		}
		old, next := facts[baseline].component(i), desired.component(i)
		node := componentNode(tree, name)
		if !old.present && !next.present {
			require.Nil(t, node, "resource %s was not applied or desired", id)
			continue
		}
		require.NotNil(t, node, id)
		requireStringFact(t, node, "type", string(old.kind), string(next.kind))
		requireStringFact(t, node, "var_name", old.value, next.value)
		require.Equal(t, old != next, node.DirectSummary().HasChanged, id)
		category := "changed"
		switch {
		case !old.present:
			category = "added"
		case !next.present:
			category = "removed"
		case old == next:
			category = "unchanged"
		}
		wantCategories[category] = append(wantCategories[category], id)
	}
	for category, entries := range map[string][]app.ComponentDiffEntry{"added": summary.Added, "removed": summary.Removed, "changed": summary.Changed, "unchanged": summary.Unchanged} {
		ids := []string{}
		for _, entry := range entries {
			ids = append(ids, entry.ComponentID)
		}
		require.Equal(t, wantCategories[category], ids, category)
	}
	after, err := json.Marshal(in)
	require.NoError(t, err)
	require.Equal(t, before, after, "composite queries must not change supplied configs or baseline maps")
	return tree
}

func TestCompositeBaselineStateMachineProperties(t *testing.T) {
	t.Parallel()
	hegel.Test(t, func(t *hegel.T) {
		n := hegel.Draw(t, hegel.Integers(3, 5))
		facts := map[string]appliedConfigFact{}
		for _, id := range []string{"cfg-a", "cfg-b", "cfg-c", "desired"} {
			facts[id] = drawAppliedConfigFact(t, n)
		}
		choices := []string{"", "cfg-a", "cfg-b", "cfg-c", "desired", "missing-config"}
		baselines := CompositeBaselines{
			Fallback: "cfg-a", Sandbox: hegel.Draw(t, hegel.SampledFrom(choices)),
			Stack: hegel.Draw(t, hegel.SampledFrom(choices)), Components: map[string]string{},
		}
		for i := range n {
			baselines.Components[fmt.Sprintf("cmp-%d", i)] = hegel.Draw(t, hegel.SampledFrom(choices))
		}
		tree := requireCompositeFacts(t, facts, baselines, n)
		order := drawShuffled(t, []int{0, 1, 2, 3, 4}[:n])
		reordered := map[string]string{}
		for _, i := range order {
			id := fmt.Sprintf("cmp-%d", i)
			reordered[id] = baselines.Components[id]
		}
		baselines.Components = reordered
		require.Equal(t, authoredObservation(tree), authoredObservation(requireCompositeFacts(t, facts, baselines, n)))
		for range hegel.Draw(t, hegel.Integers(3, 10)) {
			resource := hegel.Draw(t, hegel.Integers(0, n+1))
			baseline := hegel.Draw(t, hegel.SampledFrom(choices))
			switch resource {
			case n:
				baselines.Sandbox = baseline
			case n + 1:
				baselines.Stack = baseline
			default:
				baselines.Components[fmt.Sprintf("cmp-%d", resource)] = baseline
			}
			next := requireCompositeFacts(t, facts, baselines, n)
			for sibling := range n {
				if sibling != resource {
					name := fmt.Sprintf("resource-%d", sibling)
					require.Equal(t, authoredObservation(componentNode(tree, name)), authoredObservation(componentNode(next, name)), "only the advanced resource may change its direct observation")
				}
			}
			tree = next
		}
		for i := range n {
			baselines.Components[fmt.Sprintf("cmp-%d", i)] = "desired"
		}
		baselines.Stack, baselines.Sandbox, baselines.Fallback = "desired", "desired", "desired"
		require.False(t, requireCompositeFacts(t, facts, baselines, n).DirectSummary().HasChanged)
	})
}

func TestCompositeUniformBaselineAndInsertionProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		n := hegel.Draw(t, hegel.Integers(3, 5))
		facts := map[string]appliedConfigFact{"applied": drawAppliedConfigFact(t, n), "desired": drawAppliedConfigFact(t, n)}
		baselines := CompositeBaselines{Fallback: "applied", Stack: "applied", Sandbox: "applied", Components: map[string]string{}}
		order := drawShuffled(t, []int{0, 1, 2, 3, 4}[:n])
		for _, i := range order {
			baselines.Components[fmt.Sprintf("cmp-%d", i)] = "applied"
		}
		tree := requireCompositeFacts(t, facts, baselines, n)
		ordinary := facts["desired"].materialize().Diff(facts["applied"].materialize())
		require.Equal(t, ordinary.DirectSummary(), tree.DirectSummary())
		require.Equal(t, authoredObservation(ordinary), authoredObservation(tree))
		slices.Reverse(order)
		baselines.Components = map[string]string{}
		for _, i := range order {
			baselines.Components[fmt.Sprintf("cmp-%d", i)] = "applied"
		}
		require.Equal(t, tree, requireCompositeFacts(t, facts, baselines, n))
	})
}

func TestCompositeExplicitEmptyAndMissingBaselineProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		n := 3
		applied, desired := drawAppliedConfigFact(t, n), drawAppliedConfigFact(t, n)
		for i := range n {
			applied.components[i].present = true
		}
		desired.components[0] = applied.components[0]
		facts := map[string]appliedConfigFact{"applied": applied, "desired": desired}
		for _, explicit := range []string{"", "missing-config"} {
			baselines := CompositeBaselines{
				Fallback: "applied", Stack: "applied", Sandbox: "applied",
				Components: map[string]string{"cmp-0": explicit},
			}
			tree := requireCompositeFacts(t, facts, baselines, n)
			require.Equal(t, pkgdiff.OpAdd, topChild(componentNode(tree, "resource-0"), "var_name").Diff.Op)
			delete(baselines.Components, "cmp-0")
			tree = requireCompositeFacts(t, facts, baselines, n)
			require.Equal(t, pkgdiff.OpNoop, topChild(componentNode(tree, "resource-0"), "var_name").Diff.Op)
			baselines.Components = nil
			require.Equal(t, authoredObservation(tree), authoredObservation(requireCompositeFacts(t, facts, baselines, n)))
		}
	})
}
