package appconfiggraph

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// graphFor builds a Graph directly from connections, exercising the pure
// structural layer without any toggle/enablement overlay.
func graphFor(cccs ...*app.ComponentConfigConnection) *Graph {
	byID := make(map[string]*app.ComponentConfigConnection, len(cccs))
	for _, c := range cccs {
		byID[c.ComponentID] = c
	}
	return NewGraph(byID)
}

func TestGraph_EdgesUnionDepsAndOutputRefs(t *testing.T) {
	b := cccPlain("b", "b")
	b.Refs = []refs.Ref{{Type: refs.RefTypeComponents, Name: "a", Value: "url"}}
	g := graphFor(cccPlain("a", "a"), b, cccPlain("c", "c", "a"))

	assert.Contains(t, g.DepEdges()["b"], "a", "output ref should create a dep edge")
	assert.Contains(t, g.DepEdges()["c"], "a", "declared dependency should create a dep edge")
}

func TestGraph_DirectDependentsAndTopoSort(t *testing.T) {
	g := graphFor(
		cccPlain("a", "a"),
		cccPlain("b", "b", "a"),
		cccPlain("c", "c", "b"),
	)

	assert.Equal(t, []string{"b"}, g.DirectDependents("a"))
	assert.Equal(t, []string{"a", "b", "c"}, g.TopoSort([]string{"c", "a", "b"}))
	assert.Equal(t, []string{"c", "b", "a"}, g.ReverseTopoSort([]string{"c", "a", "b"}))
	assert.ElementsMatch(t, []string{"a", "b", "c"}, g.TransitiveDependentsClosure([]string{"a"}))
}
