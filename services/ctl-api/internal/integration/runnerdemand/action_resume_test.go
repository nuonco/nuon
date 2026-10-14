package runnerdemand_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	enumspb "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnectionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/actionworkflowrun"
	installactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/testworker"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/secretsmanager"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

const actionTaskQueue = "install-demand"

func TestRunnerDemandActionResume(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	var legacy *installactivities.Activities
	var v2 *statusactivities.Activities
	s, namespace := startSuite(t,
		fx.Provide(cloudconnectionshelpers.New, cloudformation.NewTemplates, secretsmanager.NewService, installactivities.New),
		fx.Populate(&legacy, &v2),
		fx.Invoke(fx.Annotate(func(db *gorm.DB) error {
			t.Cleanup(func() { require.NoError(t, db.Callback().Create().Remove("test:install-task-queue")) })
			return db.Callback().Create().Before("gorm:before_create").Register("test:install-task-queue", func(tx *gorm.DB) {
				if q, ok := tx.Statement.Dest.(*app.Queue); ok && q.Name == queuenames.InstallSignalsQueueName {
					q.Workflow.TaskQueue = actionTaskQueue
				}
			})
		}, fx.ParamTags(`name:"psql"`))),
		fx.Invoke(func(p testworker.WorkerParams, acts *installactivities.Activities) error {
			client, err := p.Tclient.GetNamespaceClient(p.Namespace)
			if err != nil {
				return err
			}
			w := worker.New(client, actionTaskQueue, worker.Options{DisableRegistrationAliasing: true, WorkflowPanicPolicy: worker.FailWorkflow})
			for _, shared := range p.SharedActs.AllActivities() {
				w.RegisterActivity(shared)
			}
			w.RegisterActivity(acts)
			for _, workflow := range p.QueueWkflows.All() {
				w.RegisterWorkflow(workflow)
			}
			for _, workflow := range p.HandlerWkflows.All() {
				w.RegisterWorkflow(workflow)
			}
			p.Lc.Append(fx.Hook{
				OnStart: func(context.Context) error { return w.Start() },
				OnStop:  func(context.Context) error { w.Stop(); return nil },
			})
			return nil
		}),
	)
	t.Run("completed_action_resume", func(t *testing.T) { testActionFinalizedResume(t, s, namespace) })
	t.Run("cancel_if_unfinished", func(t *testing.T) { testOwnerCancellation(t, s, namespace, legacy, v2) })
}

type actionSnapshot struct {
	run        app.InstallActionWorkflowRun
	iaw        app.InstallActionWorkflow
	step       app.WorkflowStep
	signal     app.QueueSignal
	logStreams []app.LogStream
	jobs       []app.RunnerJob
	executions []app.RunnerJobExecution
}

func testActionFinalizedResume(t *testing.T, s testServices, namespace string) {
	f := newFixture(t, s, namespace, 1, time.Minute)
	run, _ := f.prepareAction("#!/bin/sh\nexit 0\n", nil, namespace)
	s.Seeder.CreateServiceAccount(f.ctx, t, f.log.ID)

	var install app.Install
	require.NoError(t, s.DB.Where(app.Install{ID: f.workflow.OwnerID}).First(&install).Error)
	aw := s.Seeder.CreateActionWorkflow(f.ctx, t, install.AppID)
	awc := s.Seeder.CreateActionWorkflowConfig(f.ctx, t, install.AppID, install.AppConfigID, aw.ID)
	iaw := &app.InstallActionWorkflow{InstallID: install.ID, ActionWorkflowID: aw.ID}
	f.create(iaw)
	require.NoError(t, s.DB.Model(run).Updates(map[string]any{"install_action_workflow_id": iaw.ID, "action_workflow_config_id": awc.ID}).Error)
	require.NoError(t, s.DB.Model(f.step).Update("queue_signal", &signaldb.SignalData{Signal: &actionworkflowrun.Signal{
		InstallID: install.ID, AdhocActionRunID: run.ID, TriggerType: app.ActionWorkflowTriggerTypeManual,
	}}).Error)

	continuation, err := json.Marshal(map[string]any{
		"run":    map[string]string{"run_id": run.ID, "log_stream_id": f.log.ID},
		"action": map[string]string{"job_id": f.jobs[0].ID},
	})
	require.NoError(t, err)
	callbackName := "test:action-resume-continuation-" + f.step.ID
	require.NoError(t, s.DB.Callback().Create().Before("gorm:before_create").Register(callbackName, func(tx *gorm.DB) {
		if qs, ok := tx.Statement.Dest.(*app.QueueSignal); ok && qs.OwnerID == f.step.ID && qs.Type == actionworkflowrun.SignalType {
			qs.Continuation = continuation
			qs.Workflow.TaskQueue = actionTaskQueue
		}
	}))
	t.Cleanup(func() { require.NoError(t, s.DB.Callback().Create().Remove(callbackName)) })

	f.start()
	f.awaitAvailable(0)
	f.awaitWorkflow(app.StatusWaitingForRunner)
	f.awaitClosedHandlers()

	var execution app.RunnerJobExecution
	f.request(http.MethodPost, f.jobURL(0)+"/executions", struct{}{}, http.StatusCreated, &execution)
	base := f.jobURL(0) + "/executions/" + execution.ID
	f.request(http.MethodPatch, base, map[string]any{"status": "in-progress"}, http.StatusOK, nil)
	f.request(http.MethodPost, base+"/result", map[string]any{"success": true}, http.StatusCreated, nil)
	f.request(http.MethodPatch, base, map[string]any{"status": "finished"}, http.StatusOK, nil)
	f.awaitWorkflow(app.StatusSuccess)

	var qs app.QueueSignal
	require.NoError(t, s.DB.Where(app.QueueSignal{OwnerID: f.step.ID, OwnerType: "install_workflow_steps", Type: actionworkflowrun.SignalType}).First(&qs).Error)
	require.True(t, qs.RunnerDemand)
	awaitHandlerClosed(t, s, qs)

	finished := f.actionSnapshot(run.ID, iaw.ID, qs.ID)
	require.Equal(t, app.InstallActionRunStatusFinished, finished.run.Status)
	require.Equal(t, app.StatusSuccess, finished.signal.Status.Status)
	require.Equal(t, app.StatusSuccess, finished.step.Status.Status)
	require.NotNil(t, finished.iaw.AppConfigRef.AppliedConfigAt)
	require.Equal(t, run.ID, finished.iaw.AppConfigRef.AppliedConfigByID)
	require.Len(t, finished.logStreams, 1)
	require.Equal(t, f.log.ID, finished.logStreams[0].ID)
	require.False(t, finished.logStreams[0].Open)
	require.Len(t, finished.jobs, 1)
	require.Equal(t, f.jobs[0].ID, finished.jobs[0].ID)
	require.Equal(t, app.RunnerJobStatusFinished, finished.jobs[0].Status)
	require.Len(t, finished.executions, 1)
	var checkpoint struct {
		Action struct {
			JobID     string `json:"job_id"`
			Finalized bool   `json:"finalized"`
		} `json:"action"`
	}
	require.NoError(t, json.Unmarshal(finished.signal.Continuation, &checkpoint))
	require.Equal(t, f.jobs[0].ID, checkpoint.Action.JobID)
	require.True(t, checkpoint.Action.Finalized, "production finalization must checkpoint the action")

	for i := range 2 {
		ctx, cancel := context.WithTimeout(f.ctx, 30*time.Second)
		status, err := s.Queues.ResumeRunnerJobSignal(ctx, qs.ID)
		cancel()
		require.NoError(t, err, "resume %d", i)
		require.Equal(t, app.StatusSuccess, status, "resume %d", i)
		f.requireActionUnchanged(finished, fmt.Sprintf("ResumeRunnerJobSignal %d", i))
	}

	awaitHandlerClosed(t, s, qs)
	ctx, cancel := context.WithTimeout(f.ctx, 30*time.Second)
	defer cancel()
	handle, err := handler.UpdateWithStart(ctx, s.Temporal, &qs, handler.UpdateWithStartOptions{
		UpdateName: handler.ResumeRunnerJobUpdateName, WaitForStage: tclient.WorkflowUpdateStageCompleted,
	})
	if err == nil {
		var status app.Status
		if err = handle.Get(ctx, &status); err == nil {
			require.Equal(t, app.StatusSuccess, status)
		}
	}
	if err != nil {
		require.True(t, strings.Contains(err.Error(), "unknown update") || strings.Contains(err.Error(), "closing workflow"), "unexpected direct resume error: %v", err)
	}
	t.Logf("direct resume update on finished action signal returned: %v", err)
	f.requireActionUnchanged(finished, "direct UpdateWithStart")
	awaitHandlerClosed(t, s, qs)
	f.requireActionUnchanged(finished, "after handler drain")
}

func awaitHandlerClosed(t *testing.T, s testServices, qs app.QueueSignal) {
	t.Helper()
	require.Eventually(t, func() bool {
		resp, err := s.Temporal.DescribeWorkflowExecutionInNamespace(t.Context(), qs.Workflow.Namespace, qs.Workflow.ID, "")
		return err == nil && resp.GetWorkflowExecutionInfo().GetStatus() != enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING
	}, 2*time.Minute, 200*time.Millisecond, "action signal handler did not close")
}

func (f *fixture) actionSnapshot(runID, iawID, signalID string) actionSnapshot {
	f.t.Helper()
	var snap actionSnapshot
	db := f.s.DB
	require.NoError(f.t, db.Where(app.InstallActionWorkflowRun{ID: runID}).First(&snap.run).Error)
	require.NoError(f.t, db.Where(app.InstallActionWorkflow{ID: iawID}).First(&snap.iaw).Error)
	require.NoError(f.t, db.Where(app.WorkflowStep{ID: f.step.ID}).First(&snap.step).Error)
	require.NoError(f.t, db.Where(app.QueueSignal{ID: signalID}).First(&snap.signal).Error)
	require.NoError(f.t, db.Where(app.LogStream{OwnerID: runID, OwnerType: "install_action_workflow_runs"}).Order("created_at").Find(&snap.logStreams).Error)
	require.NoError(f.t, db.Scopes(scopes.WithDisableViews).Where(app.RunnerJob{OwnerID: runID, OwnerType: "install_action_workflow_runs"}).Order("created_at").Find(&snap.jobs).Error)
	require.NoError(f.t, db.Where(app.RunnerJobExecution{RunnerJobID: f.jobs[0].ID}).Order("created_at").Find(&snap.executions).Error)
	return snap
}

func (f *fixture) requireActionUnchanged(want actionSnapshot, label string) {
	f.t.Helper()
	got := f.actionSnapshot(want.run.ID, want.iaw.ID, want.signal.ID)
	require.Equal(f.t, want.run.Status, got.run.Status, "%s: run status", label)
	require.Equal(f.t, want.run.StatusDescription, got.run.StatusDescription, "%s: run status description", label)
	require.Equal(f.t, want.run.StatusV2.Status, got.run.StatusV2.Status, "%s: run status v2", label)
	require.Len(f.t, got.run.StatusV2.History, len(want.run.StatusV2.History), "%s: run status history", label)
	require.Equal(f.t, want.run.CompositeError, got.run.CompositeError, "%s: run composite error", label)
	require.True(f.t, want.run.UpdatedAt.Equal(got.run.UpdatedAt), "%s: run rewritten", label)
	require.Equal(f.t, want.iaw.AppConfigRef, got.iaw.AppConfigRef, "%s: AppliedConfigAt", label)
	require.Equal(f.t, want.step.Status.Status, got.step.Status.Status, "%s: step status", label)
	require.True(f.t, want.step.UpdatedAt.Equal(got.step.UpdatedAt), "%s: step rewritten", label)
	require.Equal(f.t, want.signal.Status.Status, got.signal.Status.Status, "%s: signal status", label)
	require.Equal(f.t, want.signal.ExecutionCount, got.signal.ExecutionCount, "%s: signal re-executed", label)
	require.JSONEq(f.t, string(want.signal.Continuation), string(got.signal.Continuation), "%s: checkpoint", label)
	require.Len(f.t, got.logStreams, len(want.logStreams), "%s: log streams", label)
	for i := range want.logStreams {
		require.Equal(f.t, want.logStreams[i].ID, got.logStreams[i].ID, "%s: log stream", label)
		require.Equal(f.t, want.logStreams[i].Open, got.logStreams[i].Open, "%s: log stream open", label)
		require.True(f.t, want.logStreams[i].UpdatedAt.Equal(got.logStreams[i].UpdatedAt), "%s: log stream rewritten", label)
	}
	require.Len(f.t, got.jobs, len(want.jobs), "%s: runner jobs", label)
	for i := range want.jobs {
		require.Equal(f.t, want.jobs[i].ID, got.jobs[i].ID, "%s: runner job", label)
		require.Equal(f.t, want.jobs[i].Status, got.jobs[i].Status, "%s: runner job status", label)
		require.True(f.t, want.jobs[i].UpdatedAt.Equal(got.jobs[i].UpdatedAt), "%s: runner job rewritten", label)
	}
	require.Len(f.t, got.executions, len(want.executions), "%s: executions", label)
}
