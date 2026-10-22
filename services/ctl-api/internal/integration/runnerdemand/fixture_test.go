package runnerdemand_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	enumspb "go.temporal.io/api/enums/v1"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

type fixture struct {
	t        *testing.T
	s        testServices
	ctx      context.Context
	workflow *app.Workflow
	step     *app.WorkflowStep
	group    *app.WorkflowStepGroup
	deploy   *app.InstallDeploy
	runner   *app.Runner
	jobs     []app.RunnerJob
	log      *app.LogStream
	queue    *app.Queue
	server   *httptest.Server
	public   *httptest.Server
}

func newFixture(t *testing.T, s testServices, namespace string, phases int, timeout time.Duration) *fixture {
	t.Helper()
	f := &fixture{t: t, s: s}
	ctx, account := s.Seeder.EnsureAccount(t.Context(), t)
	ctx, org := s.Seeder.EnsureOrg(ctx, t)
	require.NoError(t, s.DB.WithContext(ctx).Model(org).Update("sandbox_mode", false).Error)
	f.ctx = cctx.SetAccountContext(cctx.SetOrgContext(ctx, org), account)
	account.AllPermissions = permissions.Set{org.ID: permissions.PermissionAll}
	a := s.Seeder.CreateApp(f.ctx, t)
	cfg := s.Seeder.CreateAppConfig(f.ctx, t, a.ID)
	install := s.Seeder.CreateInstall(f.ctx, t, a)
	component := cfg.ComponentConfigConnections[0]
	build := s.Seeder.CreateComponentBuild(f.ctx, t, component.ID)
	ic := s.Seeder.CreateInstallComponent(f.ctx, t, install.ID, component.ComponentID)
	f.deploy = s.Seeder.CreateInstallDeploy(f.ctx, t, ic.ID, build.ID)
	runnerGroup := &app.RunnerGroup{OwnerID: install.ID, OwnerType: "installs", Type: app.RunnerGroupTypeInstall, Platform: app.AppRunnerTypeAWS}
	f.create(runnerGroup)
	f.create(&app.RunnerGroupSettings{RunnerGroupID: runnerGroup.ID, ContainerImageURL: "example.com/runner", ContainerImageTag: "test", RunnerAPIURL: "http://localhost", OTELCollectorConfiguration: "{}", SandboxMode: false})
	f.runner = &app.Runner{RunnerGroupID: runnerGroup.ID, Name: "test-runner", DisplayName: "Test runner", Status: app.RunnerStatusOffline, StatusDescription: "offline"}
	f.create(f.runner)
	f.log = &app.LogStream{OwnerID: f.deploy.ID, OwnerType: "install_deploys", Open: true}
	f.create(f.log)
	jobIDs := make([]string, phases)
	for i := range jobIDs {
		j := app.RunnerJob{RunnerID: f.runner.ID, OwnerID: f.deploy.ID, OwnerType: "install_deploys", LogStreamID: &f.log.ID,
			Status: app.RunnerJobStatusQueued, StatusDescription: "queued", Type: app.RunnerJobTypeHelmChartDeploy,
			Group: app.RunnerJobGroupDeploy, Operation: app.RunnerJobOperationTypeApplyPlan, MaxExecutions: 1,
			QueueTimeout: time.Hour, AvailableTimeout: time.Minute, OverallTimeout: time.Hour, ExecutionTimeout: timeout}
		f.create(&j)
		f.create(&app.RunnerJobPlan{RunnerJobID: j.ID, PlanJSON: `{"namespace":"acme","helm_chart":{"name":"api"}}`})
		f.jobs = append(f.jobs, j)
		jobIDs[i] = j.ID
	}
	f.workflow = s.Seeder.CreateWorkflow(f.ctx, t, install.ID, "test_runner_demand")
	f.group = s.Seeder.CreateWorkflowStepGroup(f.ctx, t, f.workflow.ID)
	f.step = &app.WorkflowStep{InstallWorkflowID: f.workflow.ID, OwnerID: install.ID, OwnerType: "installs", WorkflowStepGroupID: f.group.ID,
		Name: "apply prepared runner jobs", ExecutionType: app.WorkflowStepExecutionTypeSystem, Idx: 100,
		Status: app.NewCompositeStatus(f.ctx, app.StatusPending), StepTargetID: f.deploy.ID, StepTargetType: string(app.WorkflowStepTargetTypeInstallDeploys),
		QueueSignal: &signaldb.SignalData{Signal: &preparedJobsSignal{JobIDs: jobIDs}}}
	f.create(f.step)
	for _, name := range []string{queuenames.InstallWorkflowsQueueName, queuenames.InstallWorkflowStepGroupsQueueName, queuenames.InstallWorkflowStepsQueueName, queuenames.InstallSignalsQueueName, queuenames.InstallGenerateStepsQueueName} {
		q, err := s.Queues.Create(f.ctx, &queueclient.CreateQueueRequest{OwnerID: install.ID, OwnerType: "installs", Namespace: namespace, Name: name, MaxInFlight: 10, MaxDepth: 100})
		require.NoError(t, err)
		if name == queuenames.InstallWorkflowsQueueName {
			f.queue = q
		}
	}
	_, err := s.Queues.Create(f.ctx, &queueclient.CreateQueueRequest{OwnerID: f.runner.ID, OwnerType: "runners", Namespace: namespace, Name: string(app.RunnerJobGroupDeploy), MaxInFlight: 10, MaxDepth: 100})
	require.NoError(t, err)
	gin.SetMode(gin.TestMode)
	router := tests.NewTestRouter(tests.RouterOptions{L: s.L, DB: s.DB, TestOrg: org, TestAcc: account})
	require.NoError(t, s.Runners.RegisterRunnerRoutes(router))
	f.server = httptest.NewServer(router)
	t.Cleanup(f.server.Close)
	public := tests.NewTestRouter(tests.RouterOptions{L: s.L, DB: s.DB, TestOrg: org, TestAcc: account})
	require.NoError(t, s.Runners.RegisterPublicRoutes(public))
	f.public = httptest.NewServer(public)
	t.Cleanup(f.public.Close)
	return f
}

func (f *fixture) create(value any) {
	f.t.Helper()
	require.NoError(f.t, f.s.DB.WithContext(f.ctx).Create(value).Error)
}

func (f *fixture) start() {
	f.t.Helper()
	_, err := f.s.Queues.EnqueueSignal(f.ctx, &queueclient.EnqueueSignalRequest{QueueID: f.queue.ID, OwnerID: f.workflow.ID, OwnerType: "install_workflows", Signal: &executeflow.Signal{
		WorkflowID: f.workflow.ID, OwnerID: f.workflow.OwnerID, OwnerType: f.workflow.OwnerType,
		StepGroupQueueName: queuenames.InstallWorkflowStepGroupsQueueName, StepQueueName: queuenames.InstallWorkflowStepsQueueName,
		StepTargetQueueName: queuenames.InstallSignalsQueueName, GenerateStepsQueueName: queuenames.InstallGenerateStepsQueueName,
		Resident: true, ResidentIdleTimeout: time.Second,
	}})
	require.NoError(f.t, err)
}

func (f *fixture) request(method, url string, body any, status int, out any) {
	f.t.Helper()
	data, err := json.Marshal(body)
	require.NoError(f.t, err)
	req, err := http.NewRequestWithContext(f.ctx, method, url, bytes.NewReader(data))
	require.NoError(f.t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := f.server.Client().Do(req)
	require.NoError(f.t, err)
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	require.NoError(f.t, err)
	require.Equal(f.t, status, resp.StatusCode, "%s %s: %s", method, url, data)
	if out != nil {
		require.NoError(f.t, json.Unmarshal(data, out), "%s", data)
	}
}

func (f *fixture) claim(index int) app.RunnerJobExecution {
	f.t.Helper()
	var jobs []app.RunnerJob
	f.request(http.MethodGet, f.server.URL+"/v1/runners/"+f.runner.ID+"/jobs?group=deploy", nil, http.StatusOK, &jobs)
	require.Len(f.t, jobs, 1)
	require.Equal(f.t, f.jobs[index].ID, jobs[0].ID)
	var execution app.RunnerJobExecution
	f.request(http.MethodPost, f.jobURL(index)+"/executions", struct{}{}, http.StatusCreated, &execution)
	var step app.WorkflowStep
	require.NoError(f.t, f.s.DB.Where(app.WorkflowStep{ID: f.step.ID}).First(&step).Error)
	require.Equal(f.t, app.StatusInProgress, step.Status.Status)
	require.False(f.t, step.StartedAt.IsZero())
	require.Equal(f.t, f.jobs[index].ID, *step.RunnerJobID)
	var flw app.Workflow
	require.NoError(f.t, f.s.DB.Scopes(scopes.WithDisableViews).Where(app.Workflow{ID: f.workflow.ID}).First(&flw).Error)
	require.Equal(f.t, app.StatusInProgress, flw.Status.Status)
	var deploy app.InstallDeploy
	require.NoError(f.t, f.s.DB.Scopes(scopes.WithDisableViews).Where(app.InstallDeploy{ID: f.deploy.ID}).First(&deploy).Error)
	require.Equal(f.t, app.InstallDeployStatus(app.StatusInProgress), deploy.Status)
	require.Equal(f.t, app.StatusInProgress, deploy.StatusV2.Status)
	require.Equal(f.t, f.log.ID, *jobs[0].LogStreamID)
	return execution
}

func (f *fixture) jobURL(index int) string {
	return f.server.URL + "/v1/runner-jobs/" + f.jobs[index].ID
}

func (f *fixture) complete(index int, execution app.RunnerJobExecution) app.RunnerJobExecutionResult {
	f.t.Helper()
	base := f.jobURL(index) + "/executions/" + execution.ID
	f.request(http.MethodPatch, base, map[string]any{"status": "in-progress"}, http.StatusOK, nil)
	var result app.RunnerJobExecutionResult
	f.request(http.MethodPost, base+"/result", map[string]any{"success": true, "contents": "completed by simulated runner"}, http.StatusCreated, &result)
	f.request(http.MethodPatch, base, map[string]any{"status": "finished"}, http.StatusOK, nil)
	return result
}

func (f *fixture) awaitWorkflow(status app.Status) {
	f.t.Helper()
	require.Eventually(f.t, func() bool {
		var row app.Workflow
		return f.s.DB.Scopes(scopes.WithDisableViews).Where(app.Workflow{ID: f.workflow.ID}).First(&row).Error == nil && row.Status.Status == status
	}, 2*time.Minute, 100*time.Millisecond, "workflow did not reach %s", status)
}

func (f *fixture) awaitAvailable(index int) {
	f.t.Helper()
	require.Eventually(f.t, func() bool {
		var j app.RunnerJob
		return f.s.DB.Scopes(scopes.WithDisableViews).Where(app.RunnerJob{ID: f.jobs[index].ID}).First(&j).Error == nil && j.Status == app.RunnerJobStatusAvailable
	}, 2*time.Minute, 100*time.Millisecond)
}

func (f *fixture) signals() []app.QueueSignal {
	f.t.Helper()
	var rows []app.QueueSignal
	require.NoError(f.t, f.s.DB.Where(app.QueueSignal{OwnerID: f.workflow.ID}).Or(app.QueueSignal{OwnerID: f.group.ID}).Or(app.QueueSignal{OwnerID: f.step.ID}).Find(&rows).Error)
	return rows
}

func (f *fixture) awaitPark(index int) {
	f.t.Helper()
	f.awaitAvailable(index)
	f.awaitWorkflow(app.StatusWaitingForRunner)
	rows := f.signals()
	require.GreaterOrEqual(f.t, len(rows), 4)
	require.Eventually(f.t, func() bool {
		for _, row := range rows {
			resp, err := f.s.Temporal.DescribeWorkflowExecutionInNamespace(f.ctx, row.Workflow.Namespace, row.Workflow.ID, "")
			if err != nil || resp.GetWorkflowExecutionInfo().GetStatus() != enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED {
				return false
			}
		}
		return true
	}, time.Minute, 100*time.Millisecond, "parked workflow left Temporal handlers running")
	for _, row := range f.signals() {
		require.NotEqual(f.t, app.StatusError, row.Status.Status)
		for _, prior := range row.Status.History {
			require.NotEqual(f.t, app.StatusError, prior.Status)
		}
		if row.Type == preparedJobsSignalType || row.Type == executeflow.SignalType {
			require.Equal(f.t, app.StatusWaitingForRunner, row.Status.Status)
		}
	}
	var step app.WorkflowStep
	require.NoError(f.t, f.s.DB.Where(app.WorkflowStep{ID: f.step.ID}).First(&step).Error)
	require.Equal(f.t, app.StatusWaitingForRunner, step.Status.Status)
	require.True(f.t, step.StartedAt.IsZero())
	require.Equal(f.t, f.jobs[index].ID, *step.RunnerJobID)
	var group app.WorkflowStepGroup
	require.NoError(f.t, f.s.DB.Where(app.WorkflowStepGroup{ID: f.group.ID}).First(&group).Error)
	require.Equal(f.t, app.StatusWaitingForRunner, group.Status.Status)
	var deploy app.InstallDeploy
	require.NoError(f.t, f.s.DB.Scopes(scopes.WithDisableViews).Where(app.InstallDeploy{ID: f.deploy.ID}).First(&deploy).Error)
	require.Equal(f.t, app.StatusWaitingForRunner, deploy.StatusV2.Status)
	var runs []app.WorkflowRun
	require.NoError(f.t, f.s.DB.Where(app.WorkflowRun{WorkflowID: f.workflow.ID}).Order("created_at").Find(&runs).Error)
	require.Len(f.t, runs, index+1)
	for i, run := range runs {
		require.Equal(f.t, app.StatusWaitingForRunner, run.Status.Status)
		require.False(f.t, run.FinishedAt.IsZero())
		expected := app.WorkflowRunTypeInitial
		if i > 0 {
			expected = app.WorkflowRunTypeResume
		}
		require.Equal(f.t, expected, run.Type)
	}
	var j app.RunnerJob
	require.NoError(f.t, f.s.DB.Scopes(scopes.WithDisableViews).Where(app.RunnerJob{ID: f.jobs[index].ID}).First(&j).Error)
	require.Zero(f.t, j.QueueTimeout)
	require.Zero(f.t, j.AvailableTimeout)
	require.Zero(f.t, j.OverallTimeout)
}
