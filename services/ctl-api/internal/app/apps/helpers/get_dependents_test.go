package helpers

import (
	"context"
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

func TestGetInvertedDependentComponentsFromComponents(t *testing.T) {
	// Mock data
	ctx := context.Background()
	componentA := app.Component{ID: "cmpk8fzv4yziwnorlo9qc006zk"}
	componentB := app.Component{ID: "cmpl1d5yibj3mtculzgt914109"}
	componentC := app.Component{ID: "cmpri57fpbvr8i8up3mjcrwh1p"}
	componentD := app.Component{ID: "cmp9ctnsrgm12jxy7391tujr78", Dependencies: []*app.Component{&app.Component{ID: "cmpk8fzv4yziwnorlo9qc006zk"}, &app.Component{ID: "cmpl1d5yibj3mtculzgt914109"}, &app.Component{ID: "cmpri57fpbvr8i8up3mjcrwh1p"}}}
	componentE := app.Component{ID: "cmp6pns2xtt3toyre1y7frdz91", Dependencies: []*app.Component{&app.Component{ID: "cmpk8fzv4yziwnorlo9qc006zk"}, &app.Component{ID: "cmpl1d5yibj3mtculzgt914109"}, &app.Component{ID: "cmpri57fpbvr8i8up3mjcrwh1p"}}}
	components := []app.Component{componentA, componentB, componentC, componentD, componentE}

	helpers := &Helpers{}

	g, _, err := helpers.getInvertedDependencyGraphFromComponents(ctx, &components)

	assert.NoError(t, err)

	t.Run("should return dependent components successfully", func(t *testing.T) {
		dependentComponents, err := helpers.getInvertedDependentComponentsFromComponents(ctx, &g, &components, componentA.ID)
		assert.NoError(t, err)
		assert.Len(t, dependentComponents, 2)
		assert.Contains(t, dependentComponents, componentD)
		assert.Contains(t, dependentComponents, componentE)
	})

	t.Run("should return error if component root ID is not found", func(t *testing.T) {
		invalidRootID := "invalid"
		dependentComponents, err := helpers.getInvertedDependentComponentsFromComponents(ctx, &g, &components, invalidRootID)
		assert.Error(t, err)
		assert.Nil(t, dependentComponents)
	})
}
