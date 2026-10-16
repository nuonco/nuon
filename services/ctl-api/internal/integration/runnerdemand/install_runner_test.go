package runnerdemand_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	enumspb "go.temporal.io/api/enums/v1"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/org"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/public"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

func TestInstallRunnerDemand(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	binary := filepath.Join(t.TempDir(), "runner")
	buildCtx, cancel := context.WithTimeout(t.Context(), 5*time.Minute)
	defer cancel()
	build := exec.CommandContext(buildCtx, "go", "build", "-o", binary, "./bins/runner")
	build.Dir = "../../../../.."
	output, err := build.CombinedOutput()
	require.NoError(t, err, "%s", output)
	s, namespace := startSuite(t)
	for name, tc := range map[string]struct {
		exitCode int
		status   app.Status
	}{
		"offline_pickup_logs_and_completion": {exitCode: 0, status: app.StatusSuccess},
		"script_failure_resumes_workflow":    {exitCode: 23, status: app.StatusError},
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t, s, namespace, 1, 2*time.Minute)
			release := filepath.Join(t.TempDir(), "release")
			marker := "runner-demand-" + f.jobs[0].ID
			script := fmt.Sprintf(`#!/bin/sh
set -eu
printf '%%s\n' "$MARKER"
while [ ! -f "$RELEASE_FILE" ]; do sleep 0.1; done
printf 'answer=forty-two\n' >> "$NUON_ACTIONS_OUTPUT_FILEPATH"
exit %d
`, tc.exitCode)
			run, actionStep := f.prepareAction(script, map[string]string{"MARKER": marker, "RELEASE_FILE": release}, namespace)
			server, token := f.runnerServer()
			f.start()
			f.awaitAvailable(0)
			f.awaitWorkflow(app.StatusWaitingForRunner)
			f.awaitClosedHandlers()
			var count int64
			require.NoError(t, s.DB.Model(&app.RunnerProcess{}).Where(app.RunnerProcess{RunnerID: f.runner.ID}).Count(&count).Error)
			require.Zero(t, count)
			require.NoError(t, s.DB.Model(&app.RunnerJobExecution{}).Where(app.RunnerJobExecution{RunnerJobID: f.jobs[0].ID}).Count(&count).Error)
			require.Zero(t, count)
			var waitingStep app.WorkflowStep
			require.NoError(t, s.DB.Where(app.WorkflowStep{ID: f.step.ID}).First(&waitingStep).Error)
			require.True(t, waitingStep.StartedAt.IsZero())

			stop := startInstallRunner(t, binary, server.URL, f.runner.ID, token)
			f.awaitWorkflow(app.StatusInProgress)
			var execution app.RunnerJobExecution
			require.Eventually(t, func() bool {
				execution = app.RunnerJobExecution{}
				return s.DB.Where(app.RunnerJobExecution{RunnerJobID: f.jobs[0].ID}).First(&execution).Error == nil && execution.Status == app.RunnerJobExecutionStatusInProgress
			}, time.Minute, 100*time.Millisecond)
			var step app.WorkflowStep
			require.NoError(t, s.DB.Where(app.WorkflowStep{ID: f.step.ID}).First(&step).Error)
			require.Equal(t, app.StatusInProgress, step.Status.Status)
			require.Equal(t, f.jobs[0].ID, *step.RunnerJobID)
			require.False(t, step.StartedAt.IsZero())
			var logs []app.OtelLogRecord
			require.Eventually(t, func() bool {
				logs = nil
				return s.CHDB.Where(app.OtelLogRecord{OrgID: f.runner.OrgID, LogStreamID: f.log.ID, RunnerJobID: f.jobs[0].ID, Body: marker}).Find(&logs).Error == nil && len(logs) > 0
			}, time.Minute, 100*time.Millisecond, "runner stdout must be ingested while the action is still running")
			require.Equal(t, execution.ID, logs[0].RunnerJobExecutionID)
			require.Equal(t, "true", logs[0].LogAttributes["nuon.command_output"])
			logs = nil
			f.request(http.MethodGet, f.public.URL+"/v1/log-streams/"+f.log.ID+"/logs?body_contains="+marker, nil, http.StatusOK, &logs)
			require.NotEmpty(t, logs)
			require.Equal(t, marker, logs[0].Body)
			require.Equal(t, execution.ID, logs[0].RunnerJobExecutionID)
			var liveStep app.InstallActionWorkflowRunStep
			require.NoError(t, s.DB.Where(app.InstallActionWorkflowRunStep{ID: actionStep.ID}).First(&liveStep).Error)
			require.Equal(t, app.InstallActionWorkflowRunStepStatusInProgress, liveStep.Status)
			var liveRun app.InstallActionWorkflowRun
			require.NoError(t, s.DB.Where(app.InstallActionWorkflowRun{ID: run.ID}).First(&liveRun).Error)
			require.Equal(t, app.StatusInProgress, liveRun.StatusV2.Status)
			f.awaitClosedHandlers()

			require.NoError(t, os.WriteFile(release, nil, 0o600))
			f.awaitWorkflow(tc.status)
			var executions []app.RunnerJobExecution
			require.NoError(t, s.DB.Preload("Result").Where(app.RunnerJobExecution{RunnerJobID: f.jobs[0].ID}).Find(&executions).Error)
			require.Len(t, executions, 1)
			if tc.exitCode == 0 {
				require.Equal(t, app.RunnerJobExecutionStatusFinished, executions[0].Status)
				require.Nil(t, executions[0].Result)
				var outputs app.RunnerJobExecutionOutputs
				require.NoError(t, s.DB.Where(app.RunnerJobExecutionOutputs{RunnerJobExecutionID: execution.ID}).First(&outputs).Error)
				require.Equal(t, "forty-two", outputs.ParsedOutputs["answer"])
				require.NoError(t, s.DB.Where(app.InstallActionWorkflowRunStep{ID: actionStep.ID}).First(&liveStep).Error)
				require.Equal(t, app.InstallActionWorkflowRunStepStatusFinished, liveStep.Status)
			} else {
				require.Equal(t, app.RunnerJobExecutionStatusFailed, executions[0].Status)
				require.NotNil(t, executions[0].Result)
				require.False(t, executions[0].Result.Success)
				require.NotNil(t, executions[0].Result.ErrorMetadata["message"])
				require.Contains(t, *executions[0].Result.ErrorMetadata["message"], "exit status 23")
			}
			var steps []app.WorkflowStep
			require.NoError(t, s.DB.Where(app.WorkflowStep{InstallWorkflowID: f.workflow.ID}).Find(&steps).Error)
			require.Len(t, steps, 1)
			require.Equal(t, f.step.ID, steps[0].ID)
			require.Equal(t, tc.status, steps[0].Status.Status)
			var jobs []app.RunnerJob
			require.NoError(t, s.DB.Scopes(scopes.WithDisableViews).Where(app.RunnerJob{OwnerID: run.ID, OwnerType: "install_action_workflow_runs"}).Find(&jobs).Error)
			require.Len(t, jobs, 1)
			require.Equal(t, f.jobs[0].ID, jobs[0].ID)
			var process app.RunnerProcess
			require.Eventually(t, func() bool {
				process = app.RunnerProcess{}
				return s.DB.Where(app.RunnerProcess{RunnerID: f.runner.ID}).First(&process).Error == nil && process.ProcessStatus() == app.RunnerProcessStatusActive
			}, time.Minute, 100*time.Millisecond, "production process_init handler must run")
			require.Equal(t, app.RunnerProcessTypeInstall, process.Type)
			require.NoError(t, s.DB.Model(&app.RunnerProcess{}).Where(app.RunnerProcess{RunnerID: f.runner.ID, Type: app.RunnerProcessTypeMng}).Count(&count).Error)
			require.Zero(t, count)
			stop()
			require.NoError(t, s.DB.Where(app.RunnerProcess{ID: process.ID}).First(&process).Error)
			require.Equal(t, app.RunnerProcessStatusShutDown, process.ProcessStatus())
			t.Logf("real install runner: job=%s workflow=%s execution=%s stdout=%q outcome=%s", f.jobs[0].ID, f.workflow.ID, execution.ID, marker, tc.status)
		})
	}
}

func (f *fixture) prepareAction(script string, env map[string]string, namespace string) (*app.InstallActionWorkflowRun, *app.InstallActionWorkflowRunStep) {
	f.t.Helper()
	run := &app.InstallActionWorkflowRun{InstallID: f.workflow.OwnerID, InstallWorkflowID: &f.workflow.ID, Status: app.InstallActionRunStatusQueued, StatusV2: app.NewCompositeStatus(f.ctx, app.StatusPending), Timeout: 2 * time.Minute}
	f.create(run)
	step := &app.InstallActionWorkflowRunStep{InstallActionWorkflowRunID: run.ID, Status: app.InstallActionWorkflowRunStepStatusPending, AdHocConfig: &app.AdHocStepConfig{Name: "local-action", InlineContents: script}}
	f.create(step)
	require.NoError(f.t, f.s.DB.Model(f.step).Updates(map[string]any{"name": "run local action", "step_target_id": run.ID, "step_target_type": "install_action_workflow_runs"}).Error)
	require.NoError(f.t, f.s.DB.Model(f.log).Updates(app.LogStream{OwnerID: run.ID, OwnerType: "install_action_workflow_runs"}).Error)
	require.NoError(f.t, f.s.DB.Model(&f.jobs[0]).Updates(app.RunnerJob{OwnerID: run.ID, OwnerType: "install_action_workflow_runs", Type: app.RunnerJobTypeActionsWorkflowRun, Group: app.RunnerJobGroupActions, Operation: app.RunnerJobOperationTypeExec}).Error)
	plan := &plantypes.ActionWorkflowRunPlan{ID: run.ID, InstallID: run.InstallID, Timeout: 2 * time.Minute, Steps: []*plantypes.ActionWorkflowRunStepPlan{{ID: step.ID, InterpolatedInlineContents: script, InterpolatedEnvVars: env}}}
	data, err := json.Marshal(plan)
	require.NoError(f.t, err)
	require.NoError(f.t, f.s.DB.Model(&app.RunnerJobPlan{}).Where(app.RunnerJobPlan{RunnerJobID: f.jobs[0].ID}).Update("plan_json", string(data)).Error)
	require.NoError(f.t, f.s.DB.Model(&app.RunnerGroupSettings{}).Where(app.RunnerGroupSettings{RunnerGroupID: f.runner.RunnerGroupID}).Updates(map[string]any{
		"heart_beat_timeout": time.Second, "logging_level": "INFO", "enable_logging": true, "metadata": pgtype.Hstore{"runner.id": &f.runner.ID},
	}).Error)
	_, err = f.s.Queues.Create(f.ctx, &queueclient.CreateQueueRequest{OwnerID: f.runner.ID, OwnerType: "runners", Namespace: namespace, Name: string(app.RunnerJobGroupActions), MaxInFlight: 10, MaxDepth: 100})
	require.NoError(f.t, err)
	return run, step
}

func (f *fixture) runnerServer() (*httptest.Server, string) {
	f.t.Helper()
	acct := f.s.Seeder.CreateServiceAccount(f.ctx, f.t, f.runner.ID)
	require.NoError(f.t, f.s.Authz.CreateOrgRoles(f.ctx, f.runner.OrgID))
	require.NoError(f.t, f.s.Authz.AddAccountOrgRole(f.ctx, app.RoleTypeRunner, f.runner.OrgID, acct.ID))
	token, err := f.s.Accounts.CreateToken(f.ctx, acct.Email, time.Hour)
	require.NoError(f.t, err)
	router := tests.NewTestRouter(tests.RouterOptions{L: f.s.L, DB: f.s.DB, AdditionalMiddlewares: []gin.HandlerFunc{
		public.New(f.s.L).Handler(), f.s.Auth.Handler(), org.NewRunner(org.Params{L: f.s.L, DB: f.s.DB}).Handler(),
	}})
	require.NoError(f.t, f.s.Runners.RegisterRunnerRoutes(router))
	require.NoError(f.t, f.s.Actions.RegisterRunnerRoutes(router))
	server := httptest.NewServer(router)
	f.t.Cleanup(server.Close)
	resp, err := server.Client().Get(server.URL + "/v1/runners/" + f.runner.ID + "/settings")
	require.NoError(f.t, err)
	require.NoError(f.t, resp.Body.Close())
	require.Equal(f.t, http.StatusUnauthorized, resp.StatusCode)
	return server, token.Token
}

func (f *fixture) awaitClosedHandlers() {
	f.t.Helper()
	require.Eventually(f.t, func() bool {
		for _, row := range f.signals() {
			resp, err := f.s.Temporal.DescribeWorkflowExecutionInNamespace(f.ctx, row.Workflow.Namespace, row.Workflow.ID, "")
			if err != nil || resp.GetWorkflowExecutionInfo().GetStatus() != enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED {
				return false
			}
		}
		return true
	}, time.Minute, 100*time.Millisecond)
}

func startInstallRunner(t *testing.T, binary, url, runnerID, token string) func() {
	t.Helper()
	dir := t.TempDir()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())
	outputPath := filepath.Join(dir, "runner.log")
	output, err := os.Create(outputPath)
	require.NoError(t, err)
	cmd := exec.Command(binary, "install")
	cmd.Dir = dir
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"), "HOME=" + dir, "TMPDIR=" + dir,
		"RUNNER_ID=" + runnerID, "RUNNER_API_URL=" + url, "RUNNER_API_TOKEN=" + token,
		"HOST_IP=127.0.0.1", "GIT_REF=test", "RUNNER_PLATFORM=aws",
		"REGISTRY_DIR=" + filepath.Join(dir, "registry"), fmt.Sprintf("REGISTRY_PORT=%d", port),
		"OTEL_TRACES_EXPORTER=none", "AWS_EC2_METADATA_DISABLED=true",
	}
	cmd.Stdout, cmd.Stderr = output, output
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	require.NoError(t, cmd.Start())
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	stopped := false
	stop := func() {
		t.Helper()
		if stopped {
			return
		}
		stopped = true
		_ = cmd.Process.Signal(syscall.SIGTERM)
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("install runner exited unexpectedly: %v", err)
			}
		case <-time.After(20 * time.Second):
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			<-done
			t.Error("install runner did not stop within 20 seconds")
		}
	}
	t.Cleanup(func() {
		stop()
		_ = output.Close()
		if t.Failed() {
			data, _ := os.ReadFile(outputPath)
			t.Logf("install runner output:\n%s", data)
		}
	})
	return stop
}
