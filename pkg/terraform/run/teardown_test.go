package run

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-hclog"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/pipeline"
	"github.com/nuonco/nuon/pkg/terraform/workspace"
)

const teardownStateJSON = `{
  "format_version": "1.0",
  "values": {
    "root_module": {
      "resources": [
        {"address": "aws_rds_cluster.main", "mode": "managed", "type": "aws_rds_cluster", "name": "main", "values": {"cluster_identifier": "acme-db"}},
        {"address": "aws_db_subnet_group.main", "mode": "managed", "type": "aws_db_subnet_group", "name": "main", "values": {}},
        {"address": "data.aws_caller_identity.current", "mode": "data", "type": "aws_caller_identity", "name": "current", "values": {}}
      ]
    }
  }
}`

type fakeWorkspace struct {
	workspace.Workspace

	state *tfjson.State
	plan  *tfjson.Plan

	destroyTargets  []string
	refreshOnlyPlan bool
	stateRm         []string
}

func (f *fakeWorkspace) Show(context.Context, hclog.Logger) (*tfjson.State, error) {
	return f.state, nil
}

func (f *fakeWorkspace) ShowPlan(context.Context, hclog.Logger) (*tfjson.Plan, error) {
	return f.plan, nil
}

func (f *fakeWorkspace) PlanDestroyTargets(_ context.Context, _ hclog.Logger, targets []string) ([]byte, error) {
	f.destroyTargets = targets
	return nil, nil
}

func (f *fakeWorkspace) PlanRefreshOnly(context.Context, hclog.Logger) ([]byte, error) {
	f.refreshOnlyPlan = true
	return nil, nil
}

func (f *fakeWorkspace) StateRm(_ context.Context, _ hclog.Logger, addrs []string) error {
	f.stateRm = addrs
	return nil
}

func teardownState(t *testing.T) *tfjson.State {
	t.Helper()

	var state tfjson.State
	require.NoError(t, json.Unmarshal([]byte(teardownStateJSON), &state))
	return &state
}

func newTeardownRun(t *testing.T, ws *fakeWorkspace, retain []string) *run {
	t.Helper()

	r, err := New(validator.New(),
		WithWorkspace(ws),
		WithLogger(hclog.NewNullLogger()),
		WithOutputSettings(&OutputSettings{Ignore: true}),
		WithTeardownRetain(retain),
	)
	require.NoError(t, err)
	return r
}

func TestPlanTeardown(t *testing.T) {
	tests := map[string]struct {
		retain          []string
		wantTargets     []string
		wantRefreshOnly bool
	}{
		"destroys only resources that are not retained": {
			retain:      []string{"aws_rds_cluster.main"},
			wantTargets: []string{"aws_db_subnet_group.main"},
		},
		"unmatched entries do not change the targets": {
			retain:      []string{"aws_rds_cluster.main", "aws_s3_bucket.gone"},
			wantTargets: []string{"aws_db_subnet_group.main"},
		},
		"refresh only when everything is retained": {
			retain:          []string{"aws_rds_cluster.main", "aws_db_subnet_group.main"},
			wantRefreshOnly: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ws := &fakeWorkspace{state: teardownState(t)}
			r := newTeardownRun(t, ws, tc.retain)
			ctx := context.Background()

			require.NoError(t, r.resolveTeardown(ctx, r.Log))
			_, err := r.planTeardown(ctx, r.Log)
			require.NoError(t, err)

			assert.Equal(t, tc.wantTargets, ws.destroyTargets)
			assert.Equal(t, tc.wantRefreshOnly, ws.refreshOnlyPlan)
		})
	}
}

func TestShowTeardownPlan(t *testing.T) {
	deleteChange := func(addr string) *tfjson.ResourceChange {
		return &tfjson.ResourceChange{
			Address: addr,
			Mode:    tfjson.ManagedResourceMode,
			Change:  &tfjson.Change{Actions: tfjson.Actions{tfjson.ActionDelete}},
		}
	}

	tests := map[string]struct {
		plan        *tfjson.Plan
		wantActions map[string]tfjson.Actions
		wantErr     string
	}{
		"retained resources are added as forget changes": {
			plan: &tfjson.Plan{ResourceChanges: []*tfjson.ResourceChange{deleteChange("aws_db_subnet_group.main")}},
			wantActions: map[string]tfjson.Actions{
				"aws_db_subnet_group.main": {tfjson.ActionDelete},
				"aws_rds_cluster.main":     {tfjson.ActionForget},
			},
		},
		"plan destroying a retained dependent fails": {
			plan: &tfjson.Plan{ResourceChanges: []*tfjson.ResourceChange{
				deleteChange("aws_db_subnet_group.main"),
				deleteChange("aws_rds_cluster.main"),
			}},
			wantErr: "would destroy retained resources: aws_rds_cluster.main",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ws := &fakeWorkspace{state: teardownState(t), plan: tc.plan}
			r := newTeardownRun(t, ws, []string{"aws_rds_cluster.main"})
			ctx := context.Background()

			require.NoError(t, r.resolveTeardown(ctx, r.Log))
			plan, err := r.showTeardownPlan(ctx, r.Log)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)

			actions := make(map[string]tfjson.Actions, len(plan.ResourceChanges))
			for _, rc := range plan.ResourceChanges {
				actions[rc.Address] = rc.Change.Actions
			}
			assert.Equal(t, tc.wantActions, actions)
		})
	}
}

func TestRemoveRetainedFromState(t *testing.T) {
	tests := map[string]struct {
		retain    []string
		wantRm    []string
		wantNoRun bool
	}{
		"removes resources still in state that are retained": {
			retain: []string{"aws_rds_cluster.main"},
			wantRm: []string{"aws_rds_cluster.main"},
		},
		"skips state rm when no retained resources remain": {
			retain:    []string{"aws_s3_bucket.gone"},
			wantNoRun: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ws := &fakeWorkspace{state: teardownState(t)}
			r := newTeardownRun(t, ws, tc.retain)

			require.NoError(t, r.removeRetainedFromState(context.Background(), r.Log))
			if tc.wantNoRun {
				assert.Nil(t, ws.stateRm)
				return
			}
			assert.Equal(t, tc.wantRm, ws.stateRm)
		})
	}
}

func TestTeardownPipelineSteps(t *testing.T) {
	tests := map[string]struct {
		retain       []string
		wantDestroy  []string
		wantApplyHas bool
	}{
		"without retain": {
			wantDestroy: []string{"plan destroy", "show plan"},
		},
		"with retain": {
			retain:       []string{"aws_rds_cluster.main"},
			wantDestroy:  []string{"resolve retained resources", "plan destroy", "show plan"},
			wantApplyHas: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			r := newTeardownRun(t, &fakeWorkspace{}, tc.retain)

			destroyPipe, err := r.getDestroyPlanPipeline()
			require.NoError(t, err)
			destroySteps := stepNames(destroyPipe)
			require.Equal(t, tc.wantDestroy, destroySteps[len(destroySteps)-len(tc.wantDestroy):])

			applyPipe, err := r.getApplyPlanPipeline()
			require.NoError(t, err)
			assert.Equal(t, tc.wantApplyHas, slices.Contains(stepNames(applyPipe), "remove retained resources from state"))
		})
	}
}

func stepNames(pipe *pipeline.Pipeline) []string {
	names := make([]string, 0, len(pipe.Steps))
	for _, step := range pipe.Steps {
		names = append(names, step.Name)
	}
	return names
}
