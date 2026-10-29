package diff

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"hegel.dev/go/hegel"
)

type treeModelNode struct {
	parent           int
	hasDiff, content bool
	old, new         string
	op               Op
	impacted         bool
}

type treeModel []treeModelNode

func drawTreeModel(t *hegel.T, count int) treeModel {
	model := make(treeModel, count)
	for i := range model {
		node := &model[i]
		node.parent = -1
		if i > 0 {
			node.parent = hegel.Draw(t, hegel.Integers(0, i-1))
		}
		node.hasDiff = hegel.Draw(t, hegel.Booleans())
		node.content = hegel.Draw(t, hegel.Booleans())
		node.impacted = hegel.Draw(t, hegel.Booleans())
		node.op = hegel.Draw(t, hegel.SampledFrom([]Op{OpAdd, OpRemove, OpChange, OpNoop}))
		switch node.op {
		case OpAdd:
			node.new = "new value"
		case OpRemove:
			node.old = "old value"
		case OpChange:
			node.old, node.new = "old value", "new value"
		case OpNoop:
			if hegel.Draw(t, hegel.Booleans()) {
				node.old, node.new = "same value", "same value"
			}
		}
	}
	return model
}

func treeModelKey(index int) string {
	return fmt.Sprintf("field_%02d", index)
}

func (model treeModel) tree() *Diff {
	nodes := make([]*Diff, len(model))
	for i, node := range model {
		nodes[i] = NewDiff(WithKey(treeModelKey(i)), WithResourceID(NodeID(treeModelKey(i))))
		if node.hasDiff {
			option := WithStringDiff(node.old, node.new)
			if node.content {
				option = WithContentDiff(node.old, node.new)
			}
			option(nodes[i])
		}
		if node.parent >= 0 {
			nodes[node.parent].Children = append(nodes[node.parent].Children, nodes[i])
		}
	}
	return nodes[0]
}

func (model treeModel) descendant(child, ancestor int) bool {
	for parent := model[child].parent; parent >= 0; parent = model[parent].parent {
		if parent == ancestor {
			return true
		}
	}
	return false
}

func (model treeModel) presented(index int) bool {
	for parent := model[index].parent; parent >= 0; parent = model[parent].parent {
		if model[parent].hasDiff {
			return false
		}
	}
	return true
}

func (model treeModel) changed(index int, includeImpacts bool) bool {
	node := model[index]
	return node.hasDiff && node.op != OpNoop || includeImpacts && node.impacted
}

func (model treeModel) summary(includeImpacts bool) DiffSummary {
	result := DiffSummary{}
	for i, node := range model {
		if !model.presented(i) {
			continue
		}
		if node.hasDiff {
			switch node.op {
			case OpAdd:
				result.Added++
			case OpRemove:
				result.Removed++
			case OpChange:
				result.Changed++
			case OpNoop:
				result.Unchanged++
			}
		}
		if !includeImpacts || !node.impacted || model.changed(i, false) {
			continue
		}
		covered := false
		for j := range model {
			if model.presented(j) && model.descendant(j, i) && model.changed(j, true) {
				covered = true
			}
		}
		if !covered {
			result.Changed++
		}
	}
	result.HasChanged = result.Added+result.Removed+result.Changed > 0
	return result
}

func (model treeModel) renderIncludes(index int) bool {
	if !model.presented(index) {
		return false
	}
	if model.changed(index, true) {
		return true
	}
	for j := range model {
		if model.presented(j) && model.descendant(j, index) && model.changed(j, true) {
			return true
		}
	}
	return false
}

func TestDiffTreePresentationProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		model := drawTreeModel(t, hegel.Draw(t, hegel.Integers(1, 24)))
		tree := model.tree()
		direct := tree.DirectSummary()
		reasons := []ImpactReason{{From: "source.a", Edge: EdgeReasonInputReference}, {From: "source.b", Edge: EdgeReasonSecretReference}}
		impacts := make(map[NodeID][]ImpactReason)
		for i, node := range model {
			if node.impacted {
				impacts[NodeID(treeModelKey(i))] = reasons
			}
		}
		tree.ApplyImpacts(impacts)
		pristine := model.tree()
		pristine.ApplyImpacts(impacts)
		machine, err := json.Marshal(tree)
		require.NoError(t, err)
		require.Equal(t, direct, tree.DirectSummary())
		require.Equal(t, model.summary(false), tree.DirectSummary())
		require.Equal(t, model.summary(true), tree.Summary())
		changed, full := tree.FormatChanged(""), tree.String("")
		require.Equal(t, model.summary(true).HasChanged, changed != "")
		for i, node := range model {
			label := treeModelKey(i) + ":"
			require.Equal(t, model.renderIncludes(i), strings.Contains(changed, label), "changed inclusion of %s", label)
			fullIncluded := model.presented(i) && !(node.hasDiff && !node.content && node.op == OpNoop && node.old == "" && !node.impacted)
			require.Equal(t, fullIncluded, strings.Contains(full, label), "full inclusion of %s", label)
			resource := tree.FindResource(NodeID(treeModelKey(i)))
			require.NotNil(t, resource)
			require.Equal(t, node.impacted, resource.Impacted)
			if node.hasDiff {
				require.Equal(t, node.op, resource.Diff.Op)
				if node.content {
					require.Equal(t, node.old, resource.Diff.Before)
					require.Equal(t, node.new, resource.Diff.After)
				}
			}
		}
		require.Nil(t, tree.FindResource("missing.resource"))
		require.Equal(t, changed, tree.FormatChanged(""))
		require.Equal(t, full, tree.String(""))
		require.Equal(t, pristine, tree, "observations do not mutate the tree")
		tree.ApplyImpacts(impacts)
		require.Equal(t, pristine, tree, "reapplying the same annotations is stable")
		reorderedImpacts := make(map[NodeID][]ImpactReason)
		for _, id := range drawPermutation(t, slices.Sorted(maps.Keys(impacts))) {
			reorderedImpacts[id] = impacts[id]
		}
		reordered := model.tree()
		reordered.ApplyImpacts(reorderedImpacts)
		reorderedMachine, err := json.Marshal(reordered)
		require.NoError(t, err)
		require.Equal(t, string(machine), string(reorderedMachine))
		var decoded Diff
		require.NoError(t, json.Unmarshal(machine, &decoded))
		require.Equal(t, pristine, &decoded)
		require.Equal(t, tree.DirectSummary(), decoded.DirectSummary())
		require.Equal(t, tree.Summary(), decoded.Summary())
		require.Equal(t, changed, decoded.FormatChanged(""))
		require.Equal(t, full, decoded.String(""))

		mutated := false
		for i, node := range model {
			if node.impacted {
				resource := tree.FindResource(NodeID(treeModelKey(i)))
				require.Equal(t, reasons, resource.ImpactReasons, "resource annotations do not alias each other")
				if !mutated {
					resource.ImpactReasons[0].From = "mutated.result"
					mutated = true
				}
			}
		}
		require.Equal(t, NodeID("source.a"), reasons[0].From)
		fresh := model.tree()
		fresh.ApplyImpacts(impacts)
		require.Equal(t, pristine, fresh, "results do not share reason storage")
		reasons[0].From = "mutated.input"
		require.Equal(t, pristine, fresh, "annotations snapshot caller-owned reasons")
	})
}
