package build

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	componentsworker "github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/controlplanejob"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func TestExecuteFailurePersistence(t *testing.T) {
	for name, tc := range map[string]struct {
		failAt      string
		cause       string
		persistence string
		inactive    bool
		cancelled   bool
		bounded     bool
		legacy      bool
		planningErr bool
		planningV2  bool
	}{
		"legacy preflight failure":          {legacy: true, failAt: "git", cause: "source repository unavailable"},
		"legacy child failure":              {legacy: true, failAt: "child", cause: "registry denied"},
		"legacy child cancellation":         {legacy: true, failAt: "child", cancelled: true},
		"legacy planning status failure":    {legacy: true, planningErr: true},
		"legacy planning history failure":   {legacy: true, planningErr: true, planningV2: true},
		"new planning status failure":       {planningErr: true, cause: "status database unavailable"},
		"new planning history failure":      {planningErr: true, planningV2: true, cause: "status database unavailable"},
		"git source preflight":              {failAt: "git", cause: "source repository does not exist"},
		"component lookup":                  {failAt: "component", cause: "component not found"},
		"org provisioning prerequisite":     {failAt: "orgs", cause: "org provisioning failed"},
		"component activation prerequisite": {failAt: "components", cause: "component activation failed"},
		"ECR child failure": {
			failAt: "child", cause: "InvalidIdentityToken: No OpenIDConnect provider found",
		},
		"GAR child failure": {
			failAt: "child", cause: "PERMISSION_DENIED: artifactregistry.repositories.downloadArtifacts denied",
		},
		"ACR child failure": {
			failAt: "child", cause: "UNAUTHORIZED: pull access to acme/api denied",
		},
		"missing tag": {
			failAt: "child", cause: "MANIFEST_UNKNOWN: tag v9.7.3 does not exist",
		},
		"missing image": {
			failAt: "child", cause: "NAME_UNKNOWN: repository acme/missing does not exist",
		},
		"preflight legacy status persistence retains original cause": {
			failAt: "git", cause: "source repository unavailable", persistence: "legacy",
		},
		"prerequisite v2 status persistence retains original cause": {
			failAt: "orgs", cause: "org provisioning failed", persistence: "v2",
		},
		"diagnostics are bounded and credentials redacted": {
			failAt: "child", bounded: true,
			cause: "UNAUTHORIZED: registry denied password=acme-secret access_token=acme-token " + strings.Repeat("detail ", 20000),
		},
		"cancelled child":               {failAt: "child", cancelled: true},
		"success":                       {},
		"inactive component activation": {inactive: true},
	} {
		t.Run(name, func(t *testing.T) {
			var suite testsuite.WorkflowTestSuite
			env := suite.NewTestWorkflowEnvironment()
			env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: 5 * time.Second})
			version := workflow.Version(1)
			if tc.legacy {
				version = workflow.DefaultVersion
			}
			env.OnGetVersion(controlplanejob.BuildFailurePropagationVersion, workflow.DefaultVersion, 1).Return(version)
			failure := func(stage string) error {
				if tc.failAt != stage {
					return nil
				}
				if tc.cancelled {
					return temporal.NewCanceledError()
				}
				return temporal.NewNonRetryableApplicationError(tc.cause, "registry-test", nil)
			}

			var updates []activities.UpdateBuildStatus
			capture := func(args mock.Arguments) {
				updates = append(updates, args.Get(1).(activities.UpdateBuildStatus))
			}
			var updatesV2 []statusactivities.UpdateBuildStatusV2
			captureV2 := func(args mock.Arguments) {
				updatesV2 = append(updatesV2, args.Get(1).(statusactivities.UpdateBuildStatusV2))
			}
			if tc.planningErr {
				statusFailure := temporal.NewNonRetryableApplicationError("status database unavailable", "database", nil)
				if tc.planningV2 {
					env.OnActivity((&statusactivities.Activities{}).UpdateBuildStatusV2, mock.Anything,
						mock.MatchedBy(func(req statusactivities.UpdateBuildStatusV2) bool {
							return req.Status == app.ComponentBuildStatusPlanning
						})).Run(captureV2).Return(statusFailure).Once()
				} else {
					env.OnActivity((&activities.Activities{}).UpdateBuildStatus, mock.Anything,
						mock.MatchedBy(func(req activities.UpdateBuildStatus) bool {
							return req.Status == app.ComponentBuildStatusPlanning
						})).Run(capture).Return(statusFailure).Once()
				}
			}
			if tc.persistence == "legacy" {
				env.OnActivity((&activities.Activities{}).UpdateBuildStatus, mock.Anything,
					mock.MatchedBy(func(req activities.UpdateBuildStatus) bool { return req.Status == app.ComponentBuildStatusError })).
					Run(capture).Return(temporal.NewNonRetryableApplicationError("legacy status database unavailable", "database", nil)).Maybe()
			}
			if tc.persistence == "v2" {
				env.OnActivity((&statusactivities.Activities{}).UpdateBuildStatusV2, mock.Anything,
					mock.MatchedBy(func(req statusactivities.UpdateBuildStatusV2) bool {
						return req.Status == app.ComponentBuildStatusError
					})).
					Run(captureV2).Return(temporal.NewNonRetryableApplicationError("v2 status database unavailable", "database", nil)).Maybe()
			}
			env.OnActivity((&activities.Activities{}).UpdateBuildStatus, mock.Anything, mock.Anything).
				Run(capture).Return(nil).Maybe()
			env.OnActivity((&statusactivities.Activities{}).UpdateBuildStatusV2, mock.Anything, mock.Anything).
				Run(captureV2).Return(nil).Maybe()

			component := &app.Component{ID: "component-id", OrgID: "org-id", Status: app.ComponentStatusActive}
			if tc.inactive {
				component.Status = app.ComponentStatusError
			}
			env.OnActivity((&activities.Activities{}).GetBuildGitSource, mock.Anything, mock.Anything).
				Return((*plantypes.GitSource)(nil), failure("git")).Maybe()
			env.OnActivity((&activities.Activities{}).GetComponent, mock.Anything, mock.Anything).
				Return(component, failure("component")).Maybe()
			for _, ownerType := range []string{"orgs", "components"} {
				env.OnActivity((&queueclient.Client{}).EnsureSignal, mock.Anything,
					mock.MatchedBy(func(req *queueclient.EnsureSignalRequest) bool {
						return req.OwnerType == ownerType
					})).Return(&queueclient.EnsureSignalResponse{AlreadyComplete: true}, failure(ownerType)).Maybe()
			}
			env.OnActivity((&activities.Activities{}).UpdateStatus, mock.Anything, mock.Anything).Return(nil).Maybe()
			env.OnActivity((&statusactivities.Activities{}).UpdateComponentStatusV2, mock.Anything, mock.Anything).Return(nil).Maybe()
			childReached := false
			env.OnWorkflow((&componentsworker.Workflows{}).Build, mock.Anything,
				mock.MatchedBy(func(req componentsworker.BuildRequest) bool {
					return req.ID == component.ID && req.BuildID == "build-id" && req.SandboxMode
				})).Run(func(mock.Arguments) { childReached = true }).Return(failure("child")).Maybe()

			env.ExecuteWorkflow((&Signal{ComponentID: component.ID, BuildID: "build-id", SandboxMode: true}).Execute)
			require.True(t, env.IsWorkflowCompleted())
			err := env.GetWorkflowError()
			require.NotEmpty(t, updates, "workflow error: %v", err)
			if tc.legacy {
				if tc.planningErr {
					require.NoError(t, err)
					require.True(t, childReached)
					require.Len(t, updates, 1)
					if tc.planningV2 {
						require.Len(t, updatesV2, 1)
					} else {
						require.Empty(t, updatesV2)
					}
				} else {
					require.Error(t, err)
					wantCount := 1
					if tc.failAt == "git" {
						wantCount = 2
					}
					require.Len(t, updates, wantCount)
					require.Len(t, updatesV2, wantCount)
					if tc.cancelled {
						require.True(t, temporal.IsCanceledError(err), "%v", err)
					} else {
						require.ErrorContains(t, err, tc.cause)
					}
				}
				for _, update := range updates {
					require.Empty(t, update.Error)
				}
				env.AssertExpectations(t)
				return
			}
			if tc.planningErr {
				require.False(t, childReached)
			}
			require.NotEmpty(t, updatesV2, "workflow error: %v", err)
			last := updates[len(updates)-1]
			lastV2 := updatesV2[len(updatesV2)-1]
			require.Equal(t, "build-id", last.BuildID)
			require.Equal(t, "build-id", lastV2.BuildID)

			if tc.cancelled || (tc.failAt == "" && !tc.planningErr) {
				if tc.cancelled {
					require.True(t, temporal.IsCanceledError(err), "%v", err)
				} else {
					require.NoError(t, err)
				}
				for _, update := range updates {
					require.NotEqual(t, app.ComponentBuildStatusError, update.Status)
					require.Empty(t, update.Error)
				}
				for _, update := range updatesV2 {
					require.NotEqual(t, app.ComponentBuildStatusError, update.Status)
				}
			} else {
				require.Error(t, err)
				require.Equal(t, app.ComponentBuildStatusError, last.Status)
				require.Equal(t, app.ComponentBuildStatusError, lastV2.Status)
				require.NotEmpty(t, last.Error)
				require.NotEmpty(t, last.StatusDescription)
				require.LessOrEqual(t, utf8.RuneCountInString(last.StatusDescription), 500)
				require.LessOrEqual(t, utf8.RuneCountInString(lastV2.StatusDescription), 500)
				require.LessOrEqual(t, len(last.Error), 64*1024)
				for _, wrapper := range []string{"activity error", "child workflow execution error", "workflow execution error", "ActivityType:", "WorkflowType:"} {
					require.NotContains(t, last.StatusDescription, wrapper)
					require.NotContains(t, lastV2.StatusDescription, wrapper)
				}
				if tc.bounded {
					require.Contains(t, last.StatusDescription, "UNAUTHORIZED")
					require.Contains(t, last.Error, "UNAUTHORIZED")
					for _, secret := range []string{"acme-secret", "acme-token"} {
						require.NotContains(t, last.StatusDescription+lastV2.StatusDescription+last.Error, secret)
					}
				} else {
					require.ErrorContains(t, err, tc.cause)
					require.Contains(t, last.StatusDescription, tc.cause)
					require.Contains(t, lastV2.StatusDescription, tc.cause)
					require.Contains(t, last.Error, tc.cause)
				}
				if tc.persistence != "" {
					require.ErrorContains(t, err, tc.persistence+" status database unavailable")
				}
			}
			env.AssertExpectations(t)
		})
	}
}

func TestCancelBuildStatus(t *testing.T) {
	for name, tc := range map[string]struct {
		buildID string
		legacy  bool
		failAt  string
	}{
		"queued build":           {buildID: "queued-build-id"},
		"running build":          {buildID: "running-build-id"},
		"legacy status failure":  {buildID: "build-id", legacy: true, failAt: "status"},
		"legacy history failure": {buildID: "build-id", legacy: true, failAt: "history"},
		"new status failure":     {buildID: "build-id", failAt: "status"},
		"new history failure":    {buildID: "build-id", failAt: "history"},
	} {
		t.Run(name, func(t *testing.T) {
			var suite testsuite.WorkflowTestSuite
			env := suite.NewTestWorkflowEnvironment()
			version := workflow.Version(1)
			if tc.legacy {
				version = workflow.DefaultVersion
			}
			env.OnGetVersion(controlplanejob.BuildFailurePropagationVersion, workflow.DefaultVersion, 1).Return(version)
			failure := func(stage string) error {
				if tc.failAt == stage {
					return temporal.NewNonRetryableApplicationError("status database unavailable", "database", nil)
				}
				return nil
			}
			var updates []activities.UpdateBuildStatus
			env.OnActivity((&activities.Activities{}).UpdateBuildStatus, mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) { updates = append(updates, args.Get(1).(activities.UpdateBuildStatus)) }).Return(failure("status"))
			var updatesV2 []statusactivities.UpdateBuildStatusV2
			env.OnActivity((&statusactivities.Activities{}).UpdateBuildStatusV2, mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					updatesV2 = append(updatesV2, args.Get(1).(statusactivities.UpdateBuildStatusV2))
				}).Return(failure("history")).Maybe()

			env.ExecuteWorkflow((&Signal{ComponentID: "component-id", BuildID: tc.buildID}).Cancel)
			if !tc.legacy && tc.failAt != "" {
				require.ErrorContains(t, env.GetWorkflowError(), "status database unavailable")
			} else {
				require.NoError(t, env.GetWorkflowError())
			}
			require.NotEmpty(t, updates)
			last := updates[len(updates)-1]
			require.Equal(t, tc.buildID, last.BuildID)
			require.Equal(t, app.ComponentBuildStatus(app.StatusCancelled), last.Status)
			require.Empty(t, last.Error)
			if tc.failAt == "status" {
				require.Empty(t, updatesV2)
			} else {
				require.Len(t, updatesV2, 1)
				require.Equal(t, tc.buildID, updatesV2[0].BuildID)
				require.Equal(t, last.Status, updatesV2[0].Status)
			}
			env.AssertExpectations(t)
		})
	}
}
