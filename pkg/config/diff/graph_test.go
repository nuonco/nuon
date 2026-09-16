package diff

import (
	"testing"

	"github.com/stretchr/testify/require"
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
