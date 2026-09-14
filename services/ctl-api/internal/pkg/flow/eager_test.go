package flow

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestEagerSubset(t *testing.T) {
	group := func(idx int, eager bool) *app.WorkflowStepGroup {
		return &app.WorkflowStepGroup{GroupIdx: idx, EagerExecution: eager}
	}
	step := func(name string, groupIdx int) *app.WorkflowStep {
		return &app.WorkflowStep{Name: name, GroupIdx: groupIdx}
	}

	tests := map[string]struct {
		groups     []*app.WorkflowStepGroup
		steps      []*app.WorkflowStep
		wantGroups []int
		wantSteps  []string
	}{
		"marked groups and their steps": {
			groups:     []*app.WorkflowStepGroup{group(1, true), group(2, true), group(3, false)},
			steps:      []*app.WorkflowStep{step("state", 1), step("stack", 2), step("sandbox", 3)},
			wantGroups: []int{1, 2},
			wantSteps:  []string{"state", "stack"},
		},
		"no group marked falls back to the first": {
			groups:     []*app.WorkflowStepGroup{group(1, false), group(2, false)},
			steps:      []*app.WorkflowStep{step("state", 1), step("sandbox", 2)},
			wantGroups: []int{1},
			wantSteps:  []string{"state"},
		},
		"eager group with no steps yet": {
			groups:     []*app.WorkflowStepGroup{group(1, true)},
			steps:      nil,
			wantGroups: []int{1},
			wantSteps:  nil,
		},
		"no groups": {
			groups:     nil,
			steps:      []*app.WorkflowStep{step("orphan", 1)},
			wantGroups: nil,
			wantSteps:  nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotGroups, gotSteps := EagerSubset(tt.groups, tt.steps)

			var groupIdxs []int
			for _, g := range gotGroups {
				groupIdxs = append(groupIdxs, g.GroupIdx)
			}
			var stepNames []string
			for _, s := range gotSteps {
				stepNames = append(stepNames, s.Name)
			}

			assert.Equal(t, tt.wantGroups, groupIdxs)
			assert.Equal(t, tt.wantSteps, stepNames)
		})
	}
}
