package builds

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	branchactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

func TestBuildComponentsForceSelection(t *testing.T) {
	for _, tc := range []struct {
		name          string
		force         bool
		expectEnqueue bool
	}{
		{name: "unchanged component is skipped", force: false, expectEnqueue: false},
		{name: "forced unchanged component is enqueued", force: true, expectEnqueue: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var suite testsuite.WorkflowTestSuite
			env := suite.NewTestWorkflowEnvironment()
			// loaded CI runners starve the workflow goroutine past the 1s default
			env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: time.Minute})
			sig := &Signal{RunID: "run-1"}

			env.OnActivity((*branchactivities.Activities).GetComponentByID, mock.Anything, mock.Anything, mock.Anything).
				Return(&app.Component{ID: "component-1", Name: "api"}, nil)
			if !tc.force {
				env.OnActivity((*branchactivities.Activities).CheckBuildNeeded, mock.Anything, mock.Anything, mock.Anything).
					Return(&branchactivities.CheckBuildNeededOutput{NeedsBuild: false}, nil)
			}
			if tc.expectEnqueue {
				env.OnActivity((*sharedactivities.Activities).EnqueueSignalToOwner, mock.Anything, mock.Anything, mock.Anything).
					Return((*sharedactivities.EnqueueSignalToOwnerResponse)(nil), fmt.Errorf("enqueued marker"))
			}

			env.ExecuteWorkflow(func(ctx workflow.Context) error {
				_, err := sig.buildComponents(ctx, workflow.GetLogger(ctx), &app.AppConfig{
					ComponentIDs: []string{"component-1"},
				}, "config-new", "config-old", tc.force)
				return err
			})

			if tc.expectEnqueue {
				require.ErrorContains(t, env.GetWorkflowError(), "enqueued marker")
			} else {
				require.NoError(t, env.GetWorkflowError())
			}
			env.AssertExpectations(t)
		})
	}
}

func TestComponentsAsReadyAwaitsEnqueuedBuildsAfterSelectionError(t *testing.T) {
	for _, tc := range []struct {
		name        string
		buildStatus app.ComponentBuildStatus
		wantErrs    []string
	}{
		{
			name:        "active build",
			buildStatus: app.ComponentBuildStatusActive,
			wantErrs:    []string{"component component-1: prepare source", "github unavailable"},
		},
		{
			name:        "policy failed build",
			buildStatus: app.ComponentBuildStatusPolicyFailed,
			wantErrs:    []string{"component component-1: prepare source", "github unavailable", "one or more component builds failed"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var suite testsuite.WorkflowTestSuite
			env := suite.NewTestWorkflowEnvironment()
			env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: time.Minute})
			env.SetTestTimeout(time.Minute)
			sig := &Signal{RunID: "run-1"}

			env.OnActivity((*branchactivities.Activities).GetComponentByID, mock.Anything, mock.Anything, mock.Anything).
				Return(&app.Component{}, nil)
			env.RegisterActivity(&branchactivities.Activities{})
			env.OnActivity("LoadComparisonSourceChanged", mock.Anything, mock.Anything).
				Return(&branchactivities.LoadComparisonSourceChangedOutput{}, nil)
			env.OnActivity("CheckBuildNeeded", mock.Anything, mock.Anything).
				Return(func(_ context.Context, in *branchactivities.CheckBuildNeededInput) (*branchactivities.CheckBuildNeededOutput, error) {
					if in.ComponentID == "component-1" {
						return nil, fmt.Errorf("github unavailable")
					}
					return &branchactivities.CheckBuildNeededOutput{NeedsBuild: true}, nil
				})
			env.OnActivity((*sharedactivities.Activities).EnqueueSignalToOwner, mock.Anything, mock.Anything, mock.Anything).
				Return(&sharedactivities.EnqueueSignalToOwnerResponse{}, nil)
			env.RegisterDelayedCallback(func() {
				env.SignalWorkflow(callback.SignalName("enqueue-component-2"), callback.Result{Status: "ok"})
			}, time.Second)
			env.OnActivity((*branchactivities.Activities).CheckBuildsComplete, mock.Anything, mock.Anything, mock.Anything).
				Return(&branchactivities.AwaitBuildsCompleteOutput{
					Builds:   []branchactivities.BuildResult{{BuildID: "build-2", ComponentID: "component-2", Status: string(tc.buildStatus)}},
					AllDone:  true,
					HasError: tc.buildStatus != app.ComponentBuildStatusActive,
				}, nil).Once()

			builds := []buildEntry{{ComponentID: "component-1"}, {ComponentID: "component-2"}}
			env.ExecuteWorkflow(func(ctx workflow.Context) error {
				return sig.componentsAsReady(ctx, workflow.GetLogger(ctx), builds, "config-new", "config-old", false, func() {})
			})

			err := env.GetWorkflowError()
			for _, want := range tc.wantErrs {
				require.ErrorContains(t, err, want)
			}
			require.Equal(t, "error", builds[0].Status)
			require.Equal(t, "build-2", builds[1].BuildID)
			env.AssertExpectations(t)
		})
	}
}
