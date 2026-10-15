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

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnectionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections/helpers"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/actionworkflowrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentdeployapplyplan"
	installactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/interceptors"
	"github.com/nuonco/nuon/services/ctl-api/internal/interceptors/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/testworker"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/secretsmanager"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
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
		fx.Provide(interceptors.AsInterceptor(cctx.New)),
		fx.Populate(&legacy, &v2),
		fx.Invoke(fx.Annotate(func(db *gorm.DB) error {
			t.Cleanup(func() { require.NoError(t, db.Callback().Create().Remove("test:install-task-queue")) })
			return db.Callback().Create().Before("gorm:before_create").Register("test:install-task-queue", func(tx *gorm.DB) {
				if q, ok := tx.Statement.Dest.(*app.Queue); ok && (q.Name == queuenames.InstallSignalsQueueName || q.Name == installhelpers.InstallStateManagerQueueName) {
					q.Workflow.TaskQueue = actionTaskQueue
				}
			})
		}, fx.ParamTags(`name:"psql"`))),
		fx.Invoke(func(p testworker.WorkerParams, acts *installactivities.Activities) error {
			client, err := p.Tclient.GetNamespaceClient(p.Namespace)
			if err != nil {
				return err
			}
			w := worker.New(client, actionTaskQueue, worker.Options{
				DisableRegistrationAliasing: true, WorkflowPanicPolicy: worker.FailWorkflow, Interceptors: p.Interceptors,
			})
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
	t.Run("image_dependency_mirror_and_action", func(t *testing.T) { testActionImageSequence(t, s, namespace) })
	t.Run("cancel_if_unfinished", func(t *testing.T) { testOwnerCancellation(t, s, namespace, legacy, v2) })
	for name, failure := range map[string]string{
		"prepared_apply_resume":               "",
		"failed_apply_preserves_runner_error": "registry rejected image credentials",
	} {
		t.Run(name, func(t *testing.T) { testApplyResume(t, s, namespace, failure) })
	}
}

func testActionImageSequence(t *testing.T, s testServices, namespace string) {
	f := newFixture(t, s, namespace, 1, time.Minute)
	run, actionStep := f.prepareAction("#!/bin/sh\nexit 0\n", nil, namespace)
	s.Seeder.CreateServiceAccount(f.ctx, t, f.log.ID)
	depLog := &app.LogStream{OwnerID: f.deploy.ID, OwnerType: "install_deploys", Open: true}
	f.create(depLog)
	s.Seeder.CreateServiceAccount(f.ctx, t, depLog.ID)
	require.NoError(t, s.DB.Model(&f.jobs[0]).Updates(map[string]any{
		"owner_id": f.deploy.ID, "owner_type": "install_deploys", "log_stream_id": depLog.ID,
		"type": app.RunnerJobTypeOCISync, "group": app.RunnerJobGroupSync,
	}).Error)
	var component app.InstallComponent
	require.NoError(t, s.DB.Where(app.InstallComponent{ID: f.deploy.InstallComponentID}).First(&component).Error)
	require.NoError(t, s.DB.Model(f.step).Update("queue_signal", &signaldb.SignalData{Signal: &actionworkflowrun.Signal{
		InstallID: run.InstallID, AdhocActionRunID: run.ID, TriggerType: app.ActionWorkflowTriggerTypeManual,
	}}).Error)
	_, err := s.Queues.Create(f.ctx, &queueclient.CreateQueueRequest{
		OwnerID: run.InstallID, OwnerType: "installs", Namespace: namespace,
		Name: installhelpers.InstallStateManagerQueueName, MaxInFlight: 1, MaxDepth: 100,
	})
	require.NoError(t, err)
	_, err = s.Queues.Create(f.ctx, &queueclient.CreateQueueRequest{
		OwnerID: f.runner.ID, OwnerType: "runners", Namespace: namespace,
		Name: string(app.RunnerJobGroupSync), MaxInFlight: 10, MaxDepth: 100,
	})
	require.NoError(t, err)
	var actionPlan app.RunnerJobPlan
	require.NoError(t, s.DB.Where(app.RunnerJobPlan{RunnerJobID: f.jobs[0].ID}).First(&actionPlan).Error)
	var pinned plantypes.ActionWorkflowRunPlan
	require.NoError(t, json.Unmarshal([]byte(actionPlan.PlanJSON), &pinned))
	pinned.SourceImage = "docker.io/acme/tools:v1"
	pinned.ImageRegistry = &configs.OCIRegistryRepository{RegistryType: configs.OCIRegistryTypePublicOCI, LoginServer: "example.com", Repository: "acme/tools"}
	pinned.ImageTag = "v1"
	continuation, err := json.Marshal(map[string]any{
		"run": map[string]string{"run_id": run.ID, "log_stream_id": f.log.ID},
		"image-deps": map[string]any{"decided": true, "deps": []any{map[string]any{
			"component_id": component.ComponentID, "build_id": f.deploy.ComponentBuildID, "install_component_id": component.ID,
		}}},
		"image-deps-" + component.ComponentID:                  job.Checkpoint{JobID: f.jobs[0].ID},
		"image-deps-" + component.ComponentID + "-preparation": map[string]string{"deploy_id": f.deploy.ID, "log_stream_id": depLog.ID},
		"action-plan": &plan.CreateActionPlanResponse{Plan: &pinned},
	})
	require.NoError(t, err)
	callbackName := "test:action-image-sequence-" + f.step.ID
	require.NoError(t, s.DB.Callback().Create().Before("gorm:before_create").Register(callbackName, func(tx *gorm.DB) {
		if qs, ok := tx.Statement.Dest.(*app.QueueSignal); ok && qs.OwnerID == f.step.ID && qs.Type == actionworkflowrun.SignalType {
			qs.Continuation = continuation
			qs.Workflow.TaskQueue = actionTaskQueue
		}
	}))
	t.Cleanup(func() { require.NoError(t, s.DB.Callback().Create().Remove(callbackName)) })
	finishJob := func(index int, image map[string]any) {
		f.awaitAvailable(index)
		f.awaitWorkflow(app.StatusWaitingForRunner)
		f.awaitClosedHandlers()
		var execution app.RunnerJobExecution
		f.request(http.MethodPost, f.jobURL(index)+"/executions", struct{}{}, http.StatusCreated, &execution)
		if image != nil {
			f.request(http.MethodPost, f.jobURL(index)+"/executions/"+execution.ID+"/outputs", map[string]any{
				"outputs": map[string]any{"image": image},
			}, http.StatusCreated, nil)
		}
		f.complete(index, execution)
	}
	nextJob := func(typ app.RunnerJobType) app.RunnerJob {
		var j app.RunnerJob
		require.Eventually(t, func() bool {
			j = app.RunnerJob{}
			return s.DB.Scopes(scopes.WithDisableViews).Where(app.RunnerJob{
				OwnerID: run.ID, OwnerType: "install_action_workflow_runs", Type: typ, Status: app.RunnerJobStatusAvailable,
			}).First(&j).Error == nil
		}, time.Minute, 100*time.Millisecond)
		return j
	}
	depDigest := "sha256:" + strings.Repeat("a", 64)
	mirrorDigest := "sha256:" + strings.Repeat("b", 64)
	ref := "example.com/acme/tools@" + mirrorDigest
	f.start()
	finishJob(0, map[string]any{"ref": "example.com/acme/dependency@" + depDigest, "digest": depDigest, "repository": "acme/dependency", "tag": "dependency"})
	f.jobs = append(f.jobs, nextJob(app.RunnerJobTypeOCISync))
	var synced app.InstallDeploy
	require.NoError(t, s.DB.Where(app.InstallDeploy{ID: f.deploy.ID}).First(&synced).Error)
	require.Equal(t, app.InstallDeployStatusActive, synced.Status)
	finishJob(1, map[string]any{"ref": ref, "digest": mirrorDigest, "repository": "acme/tools", "tag": "v1"})
	f.jobs = append(f.jobs, nextJob(app.RunnerJobTypeActionsWorkflowRun))
	var savedPlan app.RunnerJobPlan
	require.NoError(t, s.DB.Where(app.RunnerJobPlan{RunnerJobID: f.jobs[2].ID}).First(&savedPlan).Error)
	var prepared plantypes.ActionWorkflowRunPlan
	require.NoError(t, json.Unmarshal([]byte(savedPlan.PlanJSON), &prepared))
	require.Equal(t, ref, prepared.ImageDigestRef)
	require.Equal(t, pinned.SourceImage, prepared.SourceImage)
	require.Equal(t, actionStep.ID, prepared.Steps[0].ID)
	finishJob(2, nil)
	f.awaitWorkflow(app.StatusSuccess)
	var qs app.QueueSignal
	require.NoError(t, s.DB.Where(app.QueueSignal{OwnerID: f.step.ID, Type: actionworkflowrun.SignalType}).First(&qs).Error)
	awaitHandlerClosed(t, s, qs)
	var cp struct {
		Deps struct {
			Published bool `json:"published"`
			Deps      []struct {
				BuildID string `json:"build_id"`
			} `json:"deps"`
		} `json:"image-deps"`
		Mirror job.Checkpoint                 `json:"image-mirror"`
		Action job.Checkpoint                 `json:"action"`
		Plan   *plan.CreateActionPlanResponse `json:"action-plan"`
	}
	require.NoError(t, json.Unmarshal(qs.Continuation, &cp))
	require.True(t, cp.Deps.Published)
	require.Len(t, cp.Deps.Deps, 1)
	require.Equal(t, f.deploy.ComponentBuildID, cp.Deps.Deps[0].BuildID)
	var phases map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(qs.Continuation, &phases))
	var dep job.Checkpoint
	require.NoError(t, json.Unmarshal(phases["image-deps-"+component.ComponentID], &dep))
	require.Equal(t, f.jobs[0].ID, dep.JobID)
	require.True(t, dep.Finalized)
	require.Equal(t, f.jobs[1].ID, cp.Mirror.JobID)
	require.True(t, cp.Mirror.Finalized)
	require.True(t, cp.Action.Finalized)
	require.Equal(t, pinned.SourceImage, cp.Plan.Plan.SourceImage)
	var count int64
	require.NoError(t, s.DB.Model(&app.OCIArtifact{}).Where(app.OCIArtifact{OwnerID: f.deploy.ID, OwnerType: "install_deploys"}).Count(&count).Error)
	require.EqualValues(t, 1, count)
	require.NoError(t, s.DB.Scopes(scopes.WithDisableViews).Model(&app.RunnerJob{}).Where(app.RunnerJob{OwnerID: run.ID, OwnerType: "install_action_workflow_runs"}).Count(&count).Error)
	require.EqualValues(t, 2, count)
	var closed app.LogStream
	require.NoError(t, s.DB.Where(app.LogStream{ID: depLog.ID}).First(&closed).Error)
	require.False(t, closed.Open)
	closed = app.LogStream{}
	require.NoError(t, s.DB.Where(app.LogStream{ID: f.log.ID}).First(&closed).Error)
	require.False(t, closed.Open)
}

func testApplyResume(t *testing.T, s testServices, namespace, failure string) {
	f := newFixture(t, s, namespace, 1, time.Minute)
	s.Seeder.CreateServiceAccount(f.ctx, t, f.log.ID)
	var component app.InstallComponent
	require.NoError(t, s.DB.Where(app.InstallComponent{ID: f.deploy.InstallComponentID}).First(&component).Error)
	require.NoError(t, s.DB.Scopes(scopes.WithDisableViews).Model(f.deploy).Updates(map[string]any{
		"install_workflow_id": f.workflow.ID, "status": app.InstallDeployStatusExecuting,
		"status_v2": app.NewCompositeStatus(f.ctx, app.StatusInProgress),
	}).Error)
	require.NoError(t, s.DB.Model(f.step).Update("queue_signal", &signaldb.SignalData{Signal: &componentdeployapplyplan.Signal{
		InstallID: f.workflow.OwnerID, InstallComponentID: component.ID, ComponentID: component.ComponentID, FlowID: f.workflow.ID,
	}}).Error)
	_, err := s.Queues.Create(f.ctx, &queueclient.CreateQueueRequest{
		OwnerID: f.workflow.OwnerID, OwnerType: "installs", Namespace: namespace,
		Name: installhelpers.InstallStateManagerQueueName, MaxInFlight: 1, MaxDepth: 100,
	})
	require.NoError(t, err)
	continuation, err := json.Marshal(map[string]any{"apply": map[string]any{
		"job_id": f.jobs[0].ID,
	}})
	require.NoError(t, err)
	callbackName := "test:apply-resume-continuation-" + f.step.ID
	require.NoError(t, s.DB.Callback().Create().Before("gorm:before_create").Register(callbackName, func(tx *gorm.DB) {
		if qs, ok := tx.Statement.Dest.(*app.QueueSignal); ok && qs.OwnerID == f.step.ID && qs.Type == componentdeployapplyplan.SignalType {
			qs.Continuation = continuation
			qs.Workflow.TaskQueue = actionTaskQueue
		}
	}))
	t.Cleanup(func() { require.NoError(t, s.DB.Callback().Create().Remove(callbackName)) })
	f.start()
	f.awaitPark(0)
	var waitingDeploy app.InstallDeploy
	var waitingLog app.LogStream
	require.NoError(t, s.DB.Where(app.InstallDeploy{ID: f.deploy.ID}).First(&waitingDeploy).Error)
	require.NoError(t, s.DB.Where(app.LogStream{ID: f.log.ID}).First(&waitingLog).Error)
	require.Nil(t, waitingDeploy.AppliedAt, "waiting must not finalize the deploy")
	require.True(t, waitingLog.Open, "waiting must not close the log stream")
	execution := f.claim(0)
	if failure != "" {
		f.request(http.MethodPatch, f.jobURL(0)+"/executions/"+execution.ID, map[string]any{
			"status": "failed", "status_description": failure,
		}, http.StatusOK, nil)
		f.awaitWorkflow(app.StatusFailedPendingRetry)
		var failed app.InstallDeploy
		var closed app.LogStream
		require.NoError(t, s.DB.Where(app.InstallDeploy{ID: f.deploy.ID}).First(&failed).Error)
		require.NoError(t, s.DB.Where(app.LogStream{ID: f.log.ID}).First(&closed).Error)
		require.Equal(t, app.InstallDeployStatusError, failed.Status)
		require.Contains(t, failed.StatusDescription, failure)
		require.Contains(t, failed.StatusV2.StatusHumanDescription, failure)
		require.Nil(t, failed.AppliedAt)
		require.False(t, closed.Open)
		return
	}
	f.complete(0, execution)
	f.awaitWorkflow(app.StatusSuccess)
	var deploy app.InstallDeploy
	var log app.LogStream
	var qs app.QueueSignal
	require.NoError(t, s.DB.Where(app.InstallDeploy{ID: f.deploy.ID}).First(&deploy).Error)
	require.NoError(t, s.DB.Where(app.LogStream{ID: f.log.ID}).First(&log).Error)
	require.NoError(t, s.DB.Where(app.QueueSignal{OwnerID: f.step.ID, Type: componentdeployapplyplan.SignalType}).First(&qs).Error)
	awaitHandlerClosed(t, s, qs)
	require.Equal(t, app.InstallDeployStatusActive, deploy.Status)
	require.NotNil(t, deploy.AppliedAt, "job completion must apply the deploy")
	require.False(t, log.Open)
	require.NoError(t, s.DB.Where(app.InstallComponent{ID: component.ID}).First(&component).Error)
	require.Equal(t, deploy.ID, component.AppConfigRef.AppliedConfigByID)
	var states []app.InstallState
	require.NoError(t, s.DB.Where(app.InstallState{InstallID: f.workflow.OwnerID, TriggeredByID: deploy.ID}).Find(&states).Error)
	require.NotEmpty(t, states, "apply finalization must regenerate state")
	var checkpoint struct {
		Apply job.Checkpoint `json:"apply"`
	}
	require.NoError(t, json.Unmarshal(qs.Continuation, &checkpoint))
	require.Equal(t, f.jobs[0].ID, checkpoint.Apply.JobID)
	require.True(t, checkpoint.Apply.Finalized)
	status, err := s.Queues.ResumeRunnerJobSignal(f.ctx, qs.ID)
	require.NoError(t, err)
	require.Equal(t, app.StatusSuccess, status)
	var unchanged app.InstallDeploy
	require.NoError(t, s.DB.Where(app.InstallDeploy{ID: deploy.ID}).First(&unchanged).Error)
	require.Equal(t, deploy.AppliedAt, unchanged.AppliedAt)
	var jobs int64
	require.NoError(t, s.DB.Scopes(scopes.WithDisableViews).Model(&app.RunnerJob{}).Where(app.RunnerJob{OwnerID: deploy.ID, OwnerType: "install_deploys"}).Count(&jobs).Error)
	require.EqualValues(t, 1, jobs, "resume must reuse the prepared job")
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
		"run":          map[string]string{"run_id": run.ID, "log_stream_id": f.log.ID},
		"image-deps":   map[string]bool{"decided": true, "published": true},
		"image-mirror": job.Checkpoint{Finalized: true},
		"action":       job.Checkpoint{JobID: f.jobs[0].ID},
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
