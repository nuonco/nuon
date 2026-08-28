package builds

import (
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

func TestShouldBuildSandboxOCI(t *testing.T) {
	for _, tc := range []struct {
		name            string
		featureEnabled  bool
		customerManaged bool
		expected        bool
	}{
		{name: "feature enabled", featureEnabled: true, expected: true},
		{name: "customer managed config", customerManaged: true, expected: true},
		{name: "ordinary config", expected: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var suite testsuite.WorkflowTestSuite
			env := suite.NewTestWorkflowEnvironment()
			sig := &Signal{}

			env.OnActivity((*branchactivities.Activities).OrgHasFeature, mock.Anything, mock.Anything, mock.Anything).
				Return(tc.featureEnabled, nil)
			if !tc.featureEnabled {
				env.OnActivity((*branchactivities.Activities).AppConfigUsesCustomerManaged, mock.Anything, mock.Anything, mock.Anything).
					Return(tc.customerManaged, nil)
			}

			var actual bool
			env.ExecuteWorkflow(func(ctx workflow.Context) error {
				var err error
				actual, err = sig.shouldBuildSandboxOCI(ctx, "org-1", "app-1", "config-1")
				return err
			})

			require.NoError(t, env.GetWorkflowError())
			require.Equal(t, tc.expected, actual)
			env.AssertExpectations(t)
		})
	}
}
