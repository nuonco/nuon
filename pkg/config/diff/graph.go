package diff

import "sort"

type NodeID string

type EdgeReason string

const (
	EdgeReasonComponentDependency EdgeReason = "component_dependency"
	EdgeReasonComponentReference  EdgeReason = "component_reference"
	EdgeReasonInputReference      EdgeReason = "input_reference"
	EdgeReasonInstallStackOutput  EdgeReason = "install_stack_output"
	EdgeReasonOperationRole       EdgeReason = "operation_role"
	EdgeReasonSandboxOutput       EdgeReason = "sandbox_output"
	EdgeReasonSecretReference     EdgeReason = "secret_reference"
	EdgeReasonStackRender         EdgeReason = "stack_render"
)

type Edge struct {
	Dependent  NodeID     `json:"dependent"`
	Dependency NodeID     `json:"dependency"`
	Reason     EdgeReason `json:"reason"`
}

type ImpactReason struct {
	From NodeID     `json:"from"`
	Edge EdgeReason `json:"edge"`
}

type Graph struct {
	dependents map[NodeID][]Edge
}

func NewGraph() *Graph {
	return &Graph{dependents: make(map[NodeID][]Edge)}
}

// AddDependency records that dependent consumes dependency. Changes propagate
// in the reverse direction, from dependency to dependent.
func (g *Graph) AddDependency(dependent, dependency NodeID, reason EdgeReason) {
	if g == nil || dependent == "" || dependency == "" || dependent == dependency {
		return
	}

	edges := g.dependents[dependency]
	for _, edge := range edges {
		if edge.Dependent == dependent && edge.Reason == reason {
			return
		}
	}
	g.dependents[dependency] = append(edges, Edge{
		Dependent:  dependent,
		Dependency: dependency,
		Reason:     reason,
	})
}

// Propagate returns every transitively impacted node and the immediate graph
// edge that caused each impact. Directly changed roots are not marked impacted.
func (g *Graph) Propagate(changed ...NodeID) map[NodeID][]ImpactReason {
	if g == nil {
		return nil
	}

	direct := make(map[NodeID]bool, len(changed))
	visited := make(map[NodeID]bool, len(changed))
	queue := make([]NodeID, 0, len(changed))
	for _, id := range changed {
		if id == "" || visited[id] {
			continue
		}
		direct[id] = true
		visited[id] = true
		queue = append(queue, id)
	}

	impacts := make(map[NodeID][]ImpactReason)
	for len(queue) > 0 {
		dependency := queue[0]
		queue = queue[1:]

		edges := append([]Edge(nil), g.dependents[dependency]...)
		sort.Slice(edges, func(i, j int) bool {
			if edges[i].Dependent != edges[j].Dependent {
				return edges[i].Dependent < edges[j].Dependent
			}
			return edges[i].Reason < edges[j].Reason
		})
		for _, edge := range edges {
			if !direct[edge.Dependent] {
				reason := ImpactReason{From: dependency, Edge: edge.Reason}
				if !containsImpactReason(impacts[edge.Dependent], reason) {
					impacts[edge.Dependent] = append(impacts[edge.Dependent], reason)
				}
			}
			if visited[edge.Dependent] {
				continue
			}
			visited[edge.Dependent] = true
			queue = append(queue, edge.Dependent)
		}
	}

	return impacts
}

func containsImpactReason(reasons []ImpactReason, target ImpactReason) bool {
	for _, reason := range reasons {
		if reason == target {
			return true
		}
	}
	return false
}
