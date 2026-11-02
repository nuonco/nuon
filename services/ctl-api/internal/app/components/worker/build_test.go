package worker

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	historypb "go.temporal.io/api/history/v1"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/encoding/protojson"

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
		legacy      bool
		statusError app.ComponentBuildStatus
		statusV2    bool
		wantTail    []string
		helm        bool
		cleanupErr  bool
	}{
		"legacy queued job failure": {legacy: true, failAt: "build child", cause: "registry denied", cleanupErr: true},
		"new queued job failure":    {failAt: "build child", cause: "registry denied", cleanupErr: true},
		"legacy helm policy denial": {legacy: true, helm: true, policyDeny: true, cause: "image must be signed"},
		"new helm policy denial":    {helm: true, policyDeny: true, cause: "image must be signed"},
		"legacy build child ordering": {
			legacy: true, failAt: "build child", cause: "registry denied",
			wantTail: []string{"child:build", "status:error", "history:error", "job:queued", "close"},
		},
		"new build child ordering": {
			failAt: "build child", cause: "registry denied",
			wantTail: []string{"child:build", "job:queued", "close", "status:error", "history:error"},
		},
		"legacy metadata child ordering": {
			legacy: true, failAt: "metadata child", policies: true, cause: "metadata denied",
			wantTail: []string{"child:metadata", "status:error", "history:error", "job:", "job:queued", "close"},
		},
		"new metadata child ordering": {
			failAt: "metadata child", policies: true, cause: "metadata denied",
			wantTail: []string{"child:metadata", "job:", "job:queued", "close", "status:error", "history:error"},
		},
		"legacy metadata cancellation ordering": {
			legacy: true, failAt: "metadata child", policies: true, cancelled: true,
			wantTail: []string{"child:metadata", "status:error", "history:error", "job:", "job:queued", "close"},
		},
		"new metadata cancellation ordering": {
			failAt: "metadata child", policies: true, cancelled: true,
			wantTail: []string{"child:metadata", "job:queued", "close"},
		},
		"legacy build cancellation ordering": {
			legacy: true, failAt: "build child", cancelled: true,
			wantTail: []string{"child:build", "status:error", "history:error", "job:queued", "close"},
		},
		"legacy policy denial":            {legacy: true, policies: true, policyDeny: true, cause: "image must be signed"},
		"legacy planning status failure":  {legacy: true, statusError: app.ComponentBuildStatusPlanning},
		"new planning status failure":     {statusError: app.ComponentBuildStatusPlanning, cause: "status database unavailable"},
		"legacy planning history failure": {legacy: true, statusError: app.ComponentBuildStatusPlanning, statusV2: true},
		"new planning history failure":    {statusError: app.ComponentBuildStatusPlanning, statusV2: true, cause: "status database unavailable"},
		"legacy building status failure":  {legacy: true, statusError: app.ComponentBuildStatusBuilding},
		"new building status failure":     {statusError: app.ComponentBuildStatusBuilding, cause: "status database unavailable"},
		"legacy building history failure": {legacy: true, statusError: app.ComponentBuildStatusBuilding, statusV2: true},
		"new building history failure":    {statusError: app.ComponentBuildStatusBuilding, statusV2: true, cause: "status database unavailable"},
		"legacy active status failure":    {legacy: true, statusError: app.ComponentBuildStatusActive},
		"new active status failure":       {statusError: app.ComponentBuildStatusActive, cause: "status database unavailable"},
		"legacy active history failure":   {legacy: true, statusError: app.ComponentBuildStatusActive, statusV2: true},
		"new active history failure":      {statusError: app.ComponentBuildStatusActive, statusV2: true, cause: "status database unavailable"},
		"legacy log stream creation":      {legacy: true, failAt: "log stream", cause: "log stream database unavailable"},
		"log stream creation":             {failAt: "log stream", cause: "log stream database unavailable"},
		"app lookup":                      {failAt: "app", cause: "component app not found"},
		"component lookup":                {failAt: "component", cause: "component not found"},
		"build lookup":                    {failAt: "build", cause: "build not found"},
		"inactive component":              {failAt: "inactive", cause: "component is not active"},
		"job creation":                    {failAt: "job", cause: "runner job database unavailable"},
		"policy lookup":                   {failAt: "policy lookup", cause: "policy database unavailable"},
		"policy org lookup":               {failAt: "org", cause: "org not found", policies: true},
		"metadata job creation":           {failAt: "metadata job", cause: "metadata job database unavailable", policies: true},
		"metadata plan persistence":       {failAt: "metadata plan", cause: "metadata plan storage unavailable", policies: true},
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
			if tc.legacy {
				if tc.statusError != "" {
					require.NoError(t, err)
					require.Equal(t, app.ComponentBuildStatusActive, last.Status)
					require.Contains(t, events, "child:build")
				} else if tc.cancelled {
					require.True(t, temporal.IsCanceledError(err), "%v", err)
				} else {
					require.ErrorContains(t, err, tc.cause)
				}
				require.Empty(t, last.Error)
				if tc.cleanupErr {
					require.NotContains(t, err.Error(), "job status database unavailable")
				}
				if tc.policyDeny {
					prefix := "image policy check failed:"
					if tc.helm {
						prefix = "helm policy check failed:"
					}
					require.ErrorContains(t, err, prefix)
					var applicationErr *temporal.ApplicationError
					require.True(t, errors.As(err, &applicationErr))
					require.False(t, applicationErr.NonRetryable())
				}
				env.AssertExpectations(t)
				return
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

func TestBuildLegacyHistoryReplay(t *testing.T) {
	for name, tc := range map[string]struct {
		fixture    string
		workflowID string
	}{
		"build child failure":    {fixture: "testdata/build_child_failure.json.gz", workflowID: "legacy-build_child_failure"},
		"metadata child failure": {fixture: "testdata/metadata_child_failure.json.gz", workflowID: "legacy-metadata_child_failure"},
	} {
		t.Run(name, func(t *testing.T) {
			file, err := os.Open(tc.fixture)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, file.Close()) })
			reader, err := gzip.NewReader(file)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, reader.Close()) })
			data, err := io.ReadAll(reader)
			require.NoError(t, err)
			var history historypb.History
			require.NoError(t, protojson.Unmarshal(data, &history))
			replayer := worker.NewWorkflowReplayer()
			replayer.RegisterWorkflow((&Workflows{cfg: &internal.Config{}}).Build)
			require.NoError(t, replayer.ReplayWorkflowHistoryWithOptions(nil, &history, worker.ReplayWorkflowHistoryOptions{
				OriginalExecution: workflow.Execution{ID: tc.workflowID},
			}))
		})
	}
}
