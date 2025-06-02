package helpers

import (
	"slices"
	"testing"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/stretchr/testify/assert"
)

func TestGetComponentsDependent(t *testing.T) {

	cases := []struct {
		name         string
		compRootID   string
		expectedCmps []string
		components   []app.Component
	}{
		{
			name:         "No dependencies",
			compRootID:   "comp1",
			expectedCmps: []string{},
			components: []app.Component{
				{ID: "comp1", Dependencies: []*app.Component{}},
				{ID: "comp2", Dependencies: []*app.Component{}},
				{ID: "comp3", Dependencies: []*app.Component{}},
			},
		},
		{
			name:         "Single dependency",
			compRootID:   "comp1",
			expectedCmps: []string{"comp2"},
			components: []app.Component{
				{ID: "comp1", Dependencies: []*app.Component{}},
				{ID: "comp2", Dependencies: []*app.Component{{ID: "comp1"}}},
				{ID: "comp3", Dependencies: []*app.Component{}},
			},
		},
		{
			name:         "Multiple dependencies",
			compRootID:   "comp1",
			expectedCmps: []string{"comp2", "comp3"},
			components: []app.Component{
				{ID: "comp1", Dependencies: []*app.Component{}},
				{ID: "comp2", Dependencies: []*app.Component{{ID: "comp1"}}},
				{ID: "comp3", Dependencies: []*app.Component{{ID: "comp1"}}},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &Helpers{}
			depCmps := h.GetComponentsDependents(tc.compRootID, tc.components)
			var depCmpIDs []string
			for _, cmp := range depCmps {
				depCmpIDs = append(depCmpIDs, cmp.ID)
			}

			if len(depCmpIDs) != len(tc.expectedCmps) {
				t.Errorf("expected %d dependent components, got %d", len(tc.expectedCmps), len(depCmpIDs))
			}

			assert.True(t, slices.Equal(depCmpIDs, tc.expectedCmps), "expected dependent components to match")
		})
	}
}
