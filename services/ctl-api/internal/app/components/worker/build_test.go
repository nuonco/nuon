package worker

import (
	"bytes"
	"compress/gzip"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"

	"github.com/nuonco/nuon/pkg/config"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	pkgdataconverter "github.com/nuonco/nuon/pkg/temporal/dataconverter"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/plan"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/controlplanejob"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	policyactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

func TestBuildFailurePersistence(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)
	var helmPayload bytes.Buffer
	writer := gzip.NewWriter(&helmPayload)
	_, err := writer.Write([]byte(`{"policy_input":[{}]}`))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	for name, tc := range map[string]struct {
		failAt      string
		cause       string
		policies    bool
		policyDeny  bool
		persistence string
		cancelled   bool
		bounded     bool
		statusError app.ComponentBuildStatus
		statusV2    bool
		wantTail    []string
		helm        bool
		cleanupErr  bool
	}{
		"new queued job failure": {failAt: "build child", cause: "registry denied", cleanupErr: true},
		"new helm policy denial": {helm: true, policyDeny: true, cause: "image must be signed"},
		"new build child ordering": {
			failAt: "build child", cause: "registry denied",
			wantTail: []string{"child:build", "job:queued", "close", "status:error", "history:error"},
		},
		"new metadata child ordering": {
			failAt: "metadata child", policies: true, cause: "metadata denied",
			wantTail: []string{"child:metadata", "job:", "job:queued", "close", "status:error", "history:error"},
		},
		"new metadata cancellation ordering": {
			failAt: "metadata child", policies: true, cancelled: true,
			wantTail: []string{"child:metadata", "job:queued", "close"},
		},
		"new planning status failure":  {statusError: app.ComponentBuildStatusPlanning, cause: "status database unavailable"},
		"new planning history failure": {statusError: app.ComponentBuildStatusPlanning, statusV2: true, cause: "status database unavailable"},
		"new building status failure":  {statusError: app.ComponentBuildStatusBuilding, cause: "status database unavailable"},
		"new building history failure": {statusError: app.ComponentBuildStatusBuilding, statusV2: true, cause: "status database unavailable"},
		"new active status failure":    {statusError: app.ComponentBuildStatusActive, cause: "status database unavailable"},
		"new active history failure":   {statusError: app.ComponentBuildStatusActive, statusV2: true, cause: "status database unavailable"},
		"log stream creation":          {failAt: "log stream", cause: "log stream database unavailable"},
		"app lookup":                   {failAt: "app", cause: "component app not found"},
		"component lookup":             {failAt: "component", cause: "component not found"},
		"build lookup":                 {failAt: "build", cause: "build not found"},
		"inactive component":           {failAt: "inactive", cause: "component is not active"},
		"job creation":                 {failAt: "job", cause: "runner job database unavailable"},
		"policy lookup":                {failAt: "policy lookup", cause: "policy database unavailable"},
		"policy org lookup":            {failAt: "org", cause: "org not found", policies: true},
		"metadata job creation":        {failAt: "metadata job", cause: "metadata job database unavailable", policies: true},
		"metadata plan persistence":    {failAt: "metadata plan", cause: "metadata plan storage unavailable", policies: true},
		"metadata ECR identity": {
			failAt: "metadata child", cause: "InvalidIdentityToken: No OpenIDConnect provider found", policies: true,
		},
		"metadata GAR permissions": {
			failAt: "metadata child", cause: "PERMISSION_DENIED: artifactregistry.repositories.downloadArtifacts denied", policies: true,
		},
		"metadata ACR authorization": {
			failAt: "metadata child", cause: "UNAUTHORIZED: pull access to acme/api denied", policies: true,
		},
		"metadata missing image": {
			failAt: "metadata child", cause: "NAME_UNKNOWN: repository acme/missing does not exist", policies: true,
		},
		"plan creation":    {failAt: "plan child", cause: "image repository configuration not found"},
		"plan persistence": {failAt: "save plan", cause: "runner job plan storage unavailable"},
		"build ECR identity": {
			failAt: "build child", cause: "InvalidIdentityToken: No OpenIDConnect provider found",
		},
		"build GAR permissions": {
			failAt: "build child", cause: "PERMISSION_DENIED: artifactregistry.repositories.downloadArtifacts denied",
		},
		"build ACR authorization": {
			failAt: "build child", cause: "UNAUTHORIZED: pull access to acme/api denied",
		},
		"build missing tag": {
			failAt: "build child", cause: "MANIFEST_UNKNOWN: tag v9.7.3 does not exist",
		},
		"build missing image": {
			failAt: "build child", cause: "NAME_UNKNOWN: repository acme/missing does not exist",
		},
		"legacy status persistence retains original cause": {
			failAt: "build child", cause: "InvalidIdentityToken: invalid identity token", persistence: "legacy",
		},
		"v2 status persistence retains original cause": {
			failAt: "build child", cause: "UNAUTHORIZED: registry access denied", persistence: "v2",
		},
		"diagnostics are not limited to the display description": {
			failAt: "build child", cause: "MANIFEST_UNKNOWN: " + strings.Repeat("diagnostic ", 150) + "trace-id=acme-trace",
		},
		"diagnostics are bounded and credentials redacted": {
			failAt: "build child", bounded: true,
			cause: "UNAUTHORIZED: registry denied password=acme-secret access_token=acme-token " + strings.Repeat("detail ", 20000),
		},
		"policy denial":      {policies: true, policyDeny: true, cause: "image must be signed"},
		"cancelled":          {failAt: "build child", cancelled: true},
		"metadata cancelled": {failAt: "metadata child", cancelled: true, policies: true},
		"success":            {},
	} {
		t.Run(name, func(t *testing.T) {
			var suite testsuite.WorkflowTestSuite
			env := suite.NewTestWorkflowEnvironment()
			env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: 5 * time.Second})
			if tc.helm {
				env.SetDataConverter(converter.NewCompositeDataConverter(
					converter.NewNilPayloadConverter(),
					converter.NewByteSlicePayloadConverter(),
					pkgdataconverter.NewJSONConverter(),
				))
			}
			failure := func(stage string) error {
				if tc.failAt != stage {
					return nil
				}
				if tc.cancelled {
					return temporal.NewCanceledError()
				}
				return temporal.NewNonRetryableApplicationError(tc.cause, "registry-test", nil)
			}

			var events []string
			var updates []activities.UpdateBuildStatus
			capture := func(args mock.Arguments) {
				req := args.Get(1).(activities.UpdateBuildStatus)
				updates = append(updates, req)
				events = append(events, "status:"+string(req.Status))
			}
			var updatesV2 []statusactivities.UpdateBuildStatusV2
			captureV2 := func(args mock.Arguments) {
				req := args.Get(1).(statusactivities.UpdateBuildStatusV2)
				updatesV2 = append(updatesV2, req)
				events = append(events, "history:"+string(req.Status))
			}
			if tc.statusError != "" {
				statusFailure := temporal.NewNonRetryableApplicationError("status database unavailable", "database", nil)
				if tc.statusV2 {
					env.OnActivity((&statusactivities.Activities{}).UpdateBuildStatusV2, mock.Anything,
						mock.MatchedBy(func(req statusactivities.UpdateBuildStatusV2) bool {
							return req.Status == tc.statusError && req.StatusDescription != "evaluating image policies"
						})).Run(captureV2).Return(statusFailure).Once()
				} else {
					env.OnActivity((&activities.Activities{}).UpdateBuildStatus, mock.Anything,
						mock.MatchedBy(func(req activities.UpdateBuildStatus) bool {
							return req.Status == tc.statusError && req.StatusDescription != "evaluating image policies"
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

			component := &app.Component{ID: "component-id", Name: "acme-api", OrgID: "org-id", Type: app.ComponentTypeExternalImage, Status: app.ComponentStatusActive}
			if tc.helm {
				component.Type = app.ComponentTypeHelmChart
			}
			if tc.failAt == "inactive" {
				component.Status = app.ComponentStatusError
			}
			env.OnActivity((&activities.Activities{}).CreateLogStream, mock.Anything, mock.Anything).
				Return(&app.LogStream{ID: "log-stream-id", RunnerAPIURL: server.URL}, failure("log stream")).Maybe()
			env.OnActivity((&activities.Activities{}).CloseLogStream, mock.Anything, mock.Anything).
				Run(func(mock.Arguments) { events = append(events, "close") }).Return(nil).Maybe()
			env.OnActivity((&activities.Activities{}).GetComponentApp, mock.Anything, mock.Anything).
				Return(&app.App{ID: "app-id", Name: "acme"}, failure("app")).Maybe()
			env.OnActivity((&activities.Activities{}).GetComponent, mock.Anything, mock.Anything).
				Return(component, failure("component")).Maybe()
			env.OnActivity((&activities.Activities{}).GetComponentBuild, mock.Anything, mock.Anything).
				Return(&app.ComponentBuild{ID: "build-id", OrgID: "org-id"}, failure("build")).Maybe()
			env.OnActivity((&activities.Activities{}).GetComponentBuildForPolicyEval, mock.Anything, mock.Anything).
				Return(&app.ComponentBuild{ComponentConfigConnection: app.ComponentConfigConnection{
					AppConfigID: "config-id", AppConfig: app.AppConfig{PoliciesConfig: app.AppPoliciesConfig{
						ID: "policies-id", Policies: []app.AppPolicyConfig{{
							ID: "policy-id", Type: config.AppPolicyTypeHelmChart, Engine: config.AppPolicyEngineOPA,
							Components: []string{component.Name},
						}},
					}},
				}}, nil).Maybe()
			env.OnActivity((&activities.Activities{}).GetRunnerJobExecutionByJobID, mock.Anything, mock.Anything).
				Return(&app.RunnerJobExecution{ID: "execution-id"}, nil).Maybe()
			env.OnActivity((&activities.Activities{}).GetRunnerJobExecutionResult, mock.Anything, mock.Anything).
				Return(&app.RunnerJobExecutionResult{ContentsGzip: helmPayload.Bytes()}, nil).Maybe()
			env.OnActivity((&activities.Activities{}).CreateBuildJob, mock.Anything, mock.Anything).
				Return(&app.RunnerJob{ID: "build-job-id"}, failure("job")).Maybe()
			env.OnActivity((&activities.Activities{}).CheckContainerImagePoliciesExist, mock.Anything, mock.Anything).
				Return(&activities.CheckContainerImagePoliciesExistResult{HasPolicies: tc.policies}, failure("policy lookup")).Maybe()
			env.OnActivity((&activities.Activities{}).GetOrg, mock.Anything, mock.Anything).
				Return(&app.Org{ID: "org-id"}, failure("org")).Maybe()
			env.OnActivity((&activities.Activities{}).CreateFetchImageMetadataJob, mock.Anything, mock.Anything).
				Return(&app.RunnerJob{ID: "metadata-job-id"}, failure("metadata job")).Maybe()
			env.OnActivity((&activities.Activities{}).SaveFetchImageMetadataPlan, mock.Anything, mock.Anything).
				Return(failure("metadata plan")).Maybe()
			env.OnActivity((&activities.Activities{}).GetImageMetadataFromJobResult, mock.Anything, mock.Anything).
				Return(&activities.GetImageMetadataFromJobResultResponse{}, nil).Maybe()
			env.OnActivity((&activities.Activities{}).PrepExternalImagePolicy, mock.Anything, mock.Anything).
				Return(&activities.PrepExternalImagePolicyResult{
					HasPolicies: true,
					Policies:    []activities.ExternalImagePolicyToEvaluate{{PolicyID: "policy-id"}},
				}, nil).Maybe()
			env.OnActivity((&policyactivities.Activities{}).EvaluateSinglePolicy, mock.Anything, mock.Anything).
				Return(&policyactivities.EvaluateSinglePolicyResult{
					Violations: []policyactivities.PolicyViolation{{Severity: "deny", Message: "image must be signed"}},
				}, nil).Maybe()
			env.OnActivity((&activities.Activities{}).SaveRunnerJobPlan, mock.Anything, mock.Anything).
				Return(failure("save plan")).Maybe()
			var cleanupErr error
			if tc.cleanupErr {
				cleanupErr = temporal.NewNonRetryableApplicationError("job status database unavailable", "database", nil)
			}
			env.OnActivity((&activities.Activities{}).UpdateJobStatus, mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					req := args.Get(1).(*activities.UpdateJobStatusRequest)
					events = append(events, "job:"+string(req.ExpectedStatus))
				}).Return(cleanupErr).Maybe()
			env.OnActivity((&activities.Activities{}).UpdateBuildSignatureVerification, mock.Anything, mock.Anything).Return(nil).Maybe()
			env.OnActivity((&sharedactivities.Activities{}).SendEmail, mock.Anything, mock.Anything).Return(nil).Maybe()

			env.OnWorkflow(plan.CreateComponentBuildPlan, mock.Anything, mock.Anything).
				Return(&plantypes.BuildPlan{ContainerImagePullPlan: &plantypes.ContainerImagePullPlan{
					RepoCfg: &configs.OCIRegistryRepository{RegistryType: configs.OCIRegistryTypePublicOCI},
				}}, failure("plan child")).Maybe()
			env.OnWorkflow((&controlplanejob.Workflows{}).ExecuteControlPlaneJob, mock.Anything,
				mock.MatchedBy(func(req *controlplanejob.ExecuteRequest) bool { return req.JobID == "metadata-job-id" })).
				Run(func(mock.Arguments) { events = append(events, "child:metadata") }).Return(failure("metadata child")).Maybe()
			env.OnWorkflow((&controlplanejob.Workflows{}).ExecuteControlPlaneJob, mock.Anything,
				mock.MatchedBy(func(req *controlplanejob.ExecuteRequest) bool { return req.JobID == "build-job-id" })).
				Run(func(mock.Arguments) { events = append(events, "child:build") }).Return(failure("build child")).Maybe()

			env.ExecuteWorkflow((&Workflows{cfg: &internal.Config{}}).Build, BuildRequest{ID: component.ID, BuildID: "build-id"})
			require.True(t, env.IsWorkflowCompleted())
			err := env.GetWorkflowError()
			require.NotEmpty(t, updates, "workflow error: %v", err)
			require.NotEmpty(t, updatesV2, "workflow error: %v", err)
			last := updates[len(updates)-1]
			lastV2 := updatesV2[len(updatesV2)-1]
			require.Equal(t, "build-id", last.BuildID)
			require.Equal(t, "build-id", lastV2.BuildID)

			if len(tc.wantTail) > 0 {
				index := slices.Index(events, tc.wantTail[0])
				require.NotEqual(t, -1, index, "%v", events)
				require.Equal(t, tc.wantTail, events[index:])
			}
			if tc.statusError == app.ComponentBuildStatusPlanning || tc.statusError == app.ComponentBuildStatusBuilding {
				require.NotContains(t, events, "child:build")
			}
			if tc.cancelled {
				require.True(t, temporal.IsCanceledError(err), "%v", err)
				for _, update := range updates {
					require.NotEqual(t, app.ComponentBuildStatusError, update.Status)
				}
				for _, update := range updatesV2 {
					require.NotEqual(t, app.ComponentBuildStatusError, update.Status)
				}
			} else if tc.failAt == "" && !tc.policyDeny && tc.statusError == "" {
				require.NoError(t, err)
				require.Equal(t, app.ComponentBuildStatusActive, last.Status)
				require.Equal(t, app.ComponentBuildStatusActive, lastV2.Status)
				require.Empty(t, last.Error)
			} else {
				require.Error(t, err)
				wantStatus := app.ComponentBuildStatusError
				if tc.policyDeny {
					wantStatus = app.ComponentBuildStatusPolicyFailed
					var applicationErr *temporal.ApplicationError
					require.True(t, errors.As(err, &applicationErr))
					require.Equal(t, "component-build-policy", applicationErr.Type())
					require.True(t, applicationErr.NonRetryable())
				}
				require.Equal(t, wantStatus, last.Status)
				require.Equal(t, wantStatus, lastV2.Status)
				require.NotEmpty(t, last.Error)
				require.NotEmpty(t, last.StatusDescription)
				require.LessOrEqual(t, utf8.RuneCountInString(last.StatusDescription), 500)
				require.LessOrEqual(t, utf8.RuneCountInString(lastV2.StatusDescription), 500)
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
					require.Contains(t, last.Error, tc.cause)
					if len(tc.cause) < 200 {
						require.Contains(t, last.StatusDescription, tc.cause)
						require.Contains(t, lastV2.StatusDescription, tc.cause)
					}
				}
				if tc.persistence != "" {
					require.ErrorContains(t, err, tc.persistence+" status database unavailable")
				}
				if tc.cleanupErr {
					require.ErrorContains(t, err, "job status database unavailable")
				}
			}
			env.AssertExpectations(t)
		})
	}
}

func TestBuildTerminalStatusWriteCount(t *testing.T) {
	for name, tc := range map[string]struct {
		policyDeny bool
		wantStatus app.ComponentBuildStatus
	}{
		"build child failure counts one terminal write": {wantStatus: app.ComponentBuildStatusError},
		"policy denial counts one terminal write":       {policyDeny: true, wantStatus: app.ComponentBuildStatusPolicyFailed},
	} {
		t.Run(name, func(t *testing.T) {
			var suite testsuite.WorkflowTestSuite
			env := suite.NewTestWorkflowEnvironment()
			env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: 5 * time.Second})

			var updates []activities.UpdateBuildStatus
			env.OnActivity((&activities.Activities{}).UpdateBuildStatus, mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) { updates = append(updates, args.Get(1).(activities.UpdateBuildStatus)) }).Return(nil).Maybe()
			var updatesV2 []statusactivities.UpdateBuildStatusV2
			env.OnActivity((&statusactivities.Activities{}).UpdateBuildStatusV2, mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					updatesV2 = append(updatesV2, args.Get(1).(statusactivities.UpdateBuildStatusV2))
				}).Return(nil).Maybe()

			env.OnActivity((&activities.Activities{}).CreateLogStream, mock.Anything, mock.Anything).
				Return(&app.LogStream{ID: "log-stream-id"}, nil).Maybe()
			env.OnActivity((&activities.Activities{}).CloseLogStream, mock.Anything, mock.Anything).Return(nil).Maybe()
			env.OnActivity((&activities.Activities{}).GetComponentApp, mock.Anything, mock.Anything).
				Return(&app.App{ID: "app-id", Name: "acme"}, nil).Maybe()
			env.OnActivity((&activities.Activities{}).GetComponent, mock.Anything, mock.Anything).
				Return(&app.Component{ID: "component-id", Name: "acme-api", OrgID: "org-id", Type: app.ComponentTypeExternalImage, Status: app.ComponentStatusActive}, nil).Maybe()
			env.OnActivity((&activities.Activities{}).GetComponentBuild, mock.Anything, mock.Anything).
				Return(&app.ComponentBuild{ID: "build-id", OrgID: "org-id"}, nil).Maybe()
			env.OnActivity((&activities.Activities{}).CheckContainerImagePoliciesExist, mock.Anything, mock.Anything).
				Return(&activities.CheckContainerImagePoliciesExistResult{HasPolicies: tc.policyDeny}, nil).Maybe()
			env.OnActivity((&activities.Activities{}).CreateBuildJob, mock.Anything, mock.Anything).
				Return(&app.RunnerJob{ID: "build-job-id"}, nil).Maybe()
			env.OnActivity((&activities.Activities{}).GetOrg, mock.Anything, mock.Anything).
				Return(&app.Org{ID: "org-id"}, nil).Maybe()
			env.OnActivity((&activities.Activities{}).CreateFetchImageMetadataJob, mock.Anything, mock.Anything).
				Return(&app.RunnerJob{ID: "metadata-job-id"}, nil).Maybe()
			env.OnActivity((&activities.Activities{}).SaveFetchImageMetadataPlan, mock.Anything, mock.Anything).Return(nil).Maybe()
			env.OnActivity((&activities.Activities{}).GetImageMetadataFromJobResult, mock.Anything, mock.Anything).
				Return(&activities.GetImageMetadataFromJobResultResponse{}, nil).Maybe()
			env.OnActivity((&activities.Activities{}).PrepExternalImagePolicy, mock.Anything, mock.Anything).
				Return(&activities.PrepExternalImagePolicyResult{
					HasPolicies: true,
					Policies:    []activities.ExternalImagePolicyToEvaluate{{PolicyID: "policy-id"}},
				}, nil).Maybe()
			env.OnActivity((&activities.Activities{}).SaveRunnerJobPlan, mock.Anything, mock.Anything).Return(nil).Maybe()
			env.OnActivity((&policyactivities.Activities{}).EvaluateSinglePolicy, mock.Anything, mock.Anything).
				Return(&policyactivities.EvaluateSinglePolicyResult{
					Violations: []policyactivities.PolicyViolation{{Severity: "deny", Message: "image must be signed"}},
				}, nil).Maybe()
			env.OnActivity((&sharedactivities.Activities{}).SendEmail, mock.Anything, mock.Anything).Return(nil).Maybe()

			env.OnWorkflow(plan.CreateComponentBuildPlan, mock.Anything, mock.Anything).
				Return(&plantypes.BuildPlan{ContainerImagePullPlan: &plantypes.ContainerImagePullPlan{
					RepoCfg: &configs.OCIRegistryRepository{RegistryType: configs.OCIRegistryTypePublicOCI},
				}}, nil).Maybe()
			childErr := error(nil)
			if !tc.policyDeny {
				childErr = temporal.NewNonRetryableApplicationError("registry denied", "registry-test", nil)
			}
			env.OnWorkflow((&controlplanejob.Workflows{}).ExecuteControlPlaneJob, mock.Anything, mock.Anything).
				Return(childErr).Maybe()

			env.ExecuteWorkflow((&Workflows{cfg: &internal.Config{}}).Build, BuildRequest{ID: "component-id", BuildID: "build-id"})
			require.True(t, env.IsWorkflowCompleted())
			require.Error(t, env.GetWorkflowError())

			require.NotEmpty(t, updates)
			require.Equal(t, app.ComponentBuildStatusPlanning, updates[0].Status)
			terminal, terminalV2 := 0, 0
			for _, update := range updates {
				if update.Status == tc.wantStatus {
					terminal++
				}
			}
			for _, update := range updatesV2 {
				if update.Status == tc.wantStatus {
					terminalV2++
				}
			}
			require.Equal(t, 1, terminal, "status writes: %v", updates)
			require.Equal(t, 1, terminalV2, "status history writes: %v", updatesV2)
			env.AssertExpectations(t)
		})
	}
}
