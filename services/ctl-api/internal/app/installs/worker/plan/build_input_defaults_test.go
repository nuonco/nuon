package plan

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func TestApplyBuildInputDefaults(t *testing.T) {
	buildCfg := &app.AppConfig{
		ID: "app-build",
		InputConfig: app.AppInputConfig{
			AppInputs: []app.AppInput{
				{Name: "sandbox_pool_replicas", Default: "2"},
				{Name: "root_domain"},
			},
		},
	}

	tests := map[string]struct {
		buildAppConfigID string
		preVersion       bool
		inputs           map[string]string
		wantLookup       bool
		want             map[string]string
	}{
		"newer build config fills a missing default": {
			buildAppConfigID: "app-build",
			inputs:           map[string]string{"root_domain": "example.com"},
			wantLookup:       true,
			want:             map[string]string{"root_domain": "example.com", "sandbox_pool_replicas": "2"},
		},
		"empty value gets the default": {
			buildAppConfigID: "app-build",
			inputs:           map[string]string{"sandbox_pool_replicas": ""},
			wantLookup:       true,
			want:             map[string]string{"sandbox_pool_replicas": "2"},
		},
		"install value wins over the default": {
			buildAppConfigID: "app-build",
			inputs:           map[string]string{"sandbox_pool_replicas": "6"},
			wantLookup:       true,
			want:             map[string]string{"sandbox_pool_replicas": "6"},
		},
		"build from the pinned config skips the lookup": {
			buildAppConfigID: "app-pinned",
			inputs:           map[string]string{"root_domain": "example.com"},
			want:             map[string]string{"root_domain": "example.com"},
		},
		"history recorded before the version skips the lookup": {
			buildAppConfigID: "app-build",
			preVersion:       true,
			inputs:           map[string]string{"root_domain": "example.com"},
			want:             map[string]string{"root_domain": "example.com"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var suite testsuite.WorkflowTestSuite
			env := suite.NewTestWorkflowEnvironment()
			env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: time.Minute})
			if tc.preVersion {
				env.OnGetVersion(buildInputDefaultsVersion, workflow.DefaultVersion, 1).Return(workflow.DefaultVersion)
			}
			var lookups []string
			env.RegisterActivityWithOptions(func(_ context.Context, req activities.GetAppConfigRequest) (*app.AppConfig, error) {
				lookups = append(lookups, req.ID)
				return buildCfg, nil
			}, activity.RegisterOptions{Name: "GetAppConfig"})

			install := &app.Install{AppConfigID: "app-pinned"}
			build := &app.ComponentBuild{ComponentConfigConnection: app.ComponentConfigConnection{AppConfigID: tc.buildAppConfigID}}

			env.ExecuteWorkflow(func(ctx workflow.Context) (*state.State, error) {
				st := state.New()
				st.Inputs = state.NewInputsState()
				st.Install = &state.InstallState{Inputs: map[string]string{}}
				for k, v := range tc.inputs {
					st.Inputs.Inputs[k] = v
					st.Install.Inputs[k] = v
				}
				if err := (&Planner{}).applyBuildInputDefaults(ctx, install, build, st); err != nil {
					return nil, err
				}
				return st, nil
			})

			require.True(t, env.IsWorkflowCompleted())
			require.NoError(t, env.GetWorkflowError())
			var got *state.State
			require.NoError(t, env.GetWorkflowResult(&got))
			require.Equal(t, tc.want, got.Inputs.Inputs)
			require.Equal(t, tc.want, got.Install.Inputs)
			if tc.wantLookup {
				require.Equal(t, []string{"app-build"}, lookups)
			} else {
				require.Empty(t, lookups)
			}
		})
	}
}
