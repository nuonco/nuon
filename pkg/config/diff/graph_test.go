package diff

import (
	"encoding/json"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"hegel.dev/go/hegel"
)

func TestGraphPropagatesTransitiveImpacts(t *testing.T) {
	graph := NewGraph()
	graph.AddDependency("stack", "role.maintenance", EdgeReasonStackRender)
	graph.AddDependency("component.api", "role.maintenance", EdgeReasonOperationRole)
	graph.AddDependency("component.web", "component.api", EdgeReasonComponentReference)
	graph.AddDependency("component.worker", "role.other", EdgeReasonOperationRole)

	impacts := graph.Propagate("role.maintenance")

	require.Equal(t, []ImpactReason{{
		From: "role.maintenance",
		Edge: EdgeReasonStackRender,
	}}, impacts["stack"])
	require.Equal(t, []ImpactReason{{
		From: "role.maintenance",
		Edge: EdgeReasonOperationRole,
	}}, impacts["component.api"])
	require.Equal(t, []ImpactReason{{
		From: "component.api",
		Edge: EdgeReasonComponentReference,
	}}, impacts["component.web"])
	require.NotContains(t, impacts, NodeID("role.maintenance"))
	require.NotContains(t, impacts, NodeID("component.worker"))
}

func TestGraphPropagationTerminatesOnCycles(t *testing.T) {
	graph := NewGraph()
	graph.AddDependency("component.api", "component.database", EdgeReasonComponentDependency)
	graph.AddDependency("component.database", "component.api", EdgeReasonComponentDependency)

	impacts := graph.Propagate("component.database")

	require.Equal(t, []ImpactReason{{
		From: "component.database",
		Edge: EdgeReasonComponentDependency,
	}}, impacts["component.api"])
	require.NotContains(t, impacts, NodeID("component.database"))
}

func TestImpactedDiffPreservesDirectOperation(t *testing.T) {
	component := NewDiff(
		WithKey("component.api"),
		WithResourceID("component.api"),
		WithChildren(NewDiff(WithKey("name"), WithStringDiff("api", "api"))),
	)
	root := NewDiff(WithKey("app"), WithChildren(component))

	root.ApplyImpacts(map[NodeID][]ImpactReason{
		"component.api": {{
			From: "role.maintenance",
			Edge: EdgeReasonOperationRole,
		}},
	})

	require.False(t, component.DirectSummary().HasChanged)
	require.True(t, component.Summary().HasChanged)
	require.True(t, component.Impacted)
	require.Contains(t, root.FormatChanged(""), "component.api")
	require.Contains(t, root.FormatChanged(""), "role.maintenance")
}

func TestGraphReasonOrderIndependentOfRoots(t *testing.T) {
	graph := NewGraph()
	graph.AddDependency("consumer", "a", EdgeReasonInputReference)
	graph.AddDependency("consumer", "b", EdgeReasonSecretReference)
	want := map[NodeID][]ImpactReason{
		"consumer": {
			{From: "a", Edge: EdgeReasonInputReference},
			{From: "b", Edge: EdgeReasonSecretReference},
		},
	}
	require.Equal(t, want, graph.Propagate("a", "b"))
	require.Equal(t, want, graph.Propagate("b", "a", "b"))
}

type graphModelEdge struct {
	from, to int
	reason   EdgeReason
}

func graphModelImpacts(ids []NodeID, edges []graphModelEdge, roots []int) map[NodeID][]ImpactReason {
	reachable := make([]bool, len(ids))
	direct := make([]bool, len(ids))
	for _, root := range roots {
		reachable[root], direct[root] = true, true
	}
	for pass := 0; pass < len(ids); pass++ {
		for _, edge := range edges {
			if edge.from != edge.to && reachable[edge.from] {
				reachable[edge.to] = true
			}
		}
	}
	reasons := make(map[NodeID]map[ImpactReason]bool)
	for _, edge := range edges {
		if edge.from == edge.to || !reachable[edge.from] || direct[edge.to] {
			continue
		}
		if reasons[ids[edge.to]] == nil {
			reasons[ids[edge.to]] = make(map[ImpactReason]bool)
		}
		reasons[ids[edge.to]][ImpactReason{From: ids[edge.from], Edge: edge.reason}] = true
	}
	result := make(map[NodeID][]ImpactReason)
	for _, id := range slices.Sorted(maps.Keys(reasons)) {
		incoming := reasons[id]
		for _, from := range ids {
			for _, kind := range graphReasonKinds {
				reason := ImpactReason{From: from, Edge: kind}
				if incoming[reason] {
					result[id] = append(result[id], reason)
				}
			}
		}
	}
	return result
}

var graphReasonKinds = []EdgeReason{
	EdgeReasonComponentDependency,
	EdgeReasonComponentReference,
	EdgeReasonInputReference,
	EdgeReasonInstallStackOutput,
	EdgeReasonOperationRole,
	EdgeReasonSandboxOutput,
	EdgeReasonSecretReference,
	EdgeReasonStackRender,
}

func TestGraphModelProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		count := hegel.Draw(t, hegel.Integers(0, 10))
		ids := []NodeID{"node.a", "node.b", "node.c", "node.d", "node.e", "node.f", "node.g", "node.h", "node.i", "node.j"}[:count]
		var edges []graphModelEdge
		var roots []int
		activeCount := count
		if count >= 3 {
			activeCount--
		}
		if activeCount > 0 {
			edges = hegel.Draw(t, hegel.Lists(hegel.Composite(func(t hegel.TestCase) graphModelEdge {
				return graphModelEdge{
					from:   hegel.Draw(t, hegel.Integers(0, activeCount-1)),
					to:     hegel.Draw(t, hegel.Integers(0, activeCount-1)),
					reason: hegel.Draw(t, hegel.SampledFrom(graphReasonKinds)),
				}
			})).MaxSize(35))
			roots = hegel.Draw(t, hegel.Lists(hegel.Integers(0, activeCount-1)).MaxSize(count))
		}
		if count >= 6 && hegel.Draw(t, hegel.Booleans()) {
			edges = append(edges,
				graphModelEdge{0, 2, EdgeReasonInputReference},
				graphModelEdge{0, 2, EdgeReasonStackRender},
				graphModelEdge{1, 2, EdgeReasonSecretReference},
				graphModelEdge{2, 3, EdgeReasonComponentDependency},
				graphModelEdge{3, 2, EdgeReasonComponentReference},
				graphModelEdge{0, 4, EdgeReasonOperationRole},
				graphModelEdge{2, 4, EdgeReasonSandboxOutput},
				graphModelEdge{4, 3, EdgeReasonComponentDependency},
				graphModelEdge{4, 0, EdgeReasonInstallStackOutput},
			)
		}
		if count >= 7 {
			roots = append([]int{0, 1}, roots...)
		}
		changed := []NodeID{"", "ghost.root"}
		for _, root := range roots {
			changed = append(changed, ids[root], ids[root])
		}
		makeGraph := func(order []graphModelEdge) *Graph {
			graph := NewGraph()
			for _, edge := range order {
				graph.AddDependency(ids[edge.to], ids[edge.from], edge.reason)
				graph.AddDependency(ids[edge.to], ids[edge.from], edge.reason)
			}
			if count > 0 {
				graph.AddDependency("", ids[0], EdgeReasonStackRender)
				graph.AddDependency(ids[count-1], "", EdgeReasonStackRender)
			}
			return graph
		}
		graph := makeGraph(edges)
		rootSnapshot := slices.Clone(changed)
		edgeSnapshot := make(map[NodeID][]Edge)
		for _, id := range slices.Sorted(maps.Keys(graph.dependents)) {
			edgeSnapshot[id] = slices.Clone(graph.dependents[id])
		}
		want := graphModelImpacts(ids, edges, roots)
		got := graph.Propagate(changed...)
		require.Equal(t, want, got)
		if count >= 3 {
			require.NotContains(t, got, ids[count-1], "disconnected control")
		}
		require.Empty(t, graph.Propagate())
		permuted := makeGraph(drawPermutation(t, edges))
		permutedRoots := drawPermutation(t, changed)
		require.Equal(t, want, permuted.Propagate(permutedRoots...))
		require.Equal(t, want, graph.Propagate(changed...))

		if count > 0 {
			extra := hegel.Draw(t, hegel.Integers(0, count-1))
			moreRoots := append(append([]int(nil), roots...), extra)
			more := graph.Propagate(append(append([]NodeID(nil), changed...), ids[extra])...)
			require.Equal(t, graphModelImpacts(ids, edges, moreRoots), more)
			for _, id := range slices.Sorted(maps.Keys(got)) {
				if id != ids[extra] {
					require.Contains(t, more, id, "combined affected set cannot shrink")
				}
			}
		}

		model := drawTreeModel(t, count+1)
		model[0].hasDiff, model[0].impacted = false, false
		direct := make([]bool, count)
		for _, root := range roots {
			direct[root] = true
		}
		for i, id := range ids {
			node := &model[i+1]
			node.impacted = len(want[id]) > 0
			if direct[i] {
				node.hasDiff = true
				if node.op == OpNoop {
					node.op, node.old, node.new = OpChange, "old value", "new value"
				}
			} else {
				node.op, node.old, node.new = OpNoop, "same value", "same value"
			}
		}
		makeTree := func(graph *Graph, roots []NodeID) *Diff {
			tree := model.tree()
			for i, id := range ids {
				tree.FindResource(NodeID(treeModelKey(i + 1))).ResourceID = id
			}
			tree.ApplyImpacts(graph.Propagate(roots...))
			return tree
		}
		tree, reorderedTree := makeTree(graph, changed), makeTree(permuted, permutedRoots)
		require.Equal(t, model.summary(false), tree.DirectSummary())
		require.Equal(t, model.summary(true), tree.Summary())
		require.Equal(t, model.summary(true).HasChanged, tree.FormatChanged("") != "")
		for i, id := range ids {
			resource := tree.FindResource(id)
			require.NotNil(t, resource)
			require.Equal(t, model[i+1].impacted, resource.Impacted)
			require.Equal(t, want[id], resource.ImpactReasons)
			if model[i+1].hasDiff {
				require.Equal(t, model[i+1].op, resource.Diff.Op)
			}
			require.Equal(t, model.renderIncludes(i+1), strings.Contains(tree.FormatChanged(""), treeModelKey(i+1)+":"))
			if model[i+1].impacted && model[i+1].op == OpNoop {
				for _, reason := range want[id] {
					require.Contains(t, resource.FormatChanged(""), string(reason.From)+" via "+string(reason.Edge))
				}
			}
		}
		machine, err := json.Marshal(tree)
		require.NoError(t, err)
		reorderedMachine, err := json.Marshal(reorderedTree)
		require.NoError(t, err)
		require.Equal(t, string(machine), string(reorderedMachine))
		require.Equal(t, tree.FormatChanged(""), reorderedTree.FormatChanged(""))
		require.Equal(t, tree.String(""), reorderedTree.String(""))

		independent := graph.Propagate(changed...)
		for _, id := range slices.Sorted(maps.Keys(got)) {
			got[id][0] = ImpactReason{From: "mutated", Edge: EdgeReasonStackRender}
			delete(got, id)
		}
		require.Equal(t, want, independent)
		require.Equal(t, want, graph.Propagate(changed...))
		require.Equal(t, rootSnapshot, changed)
		require.Equal(t, edgeSnapshot, graph.dependents)
		graph.AddDependency("new.consumer", "ghost.root", EdgeReasonStackRender)
		require.Equal(t, want, independent)
	})
}
