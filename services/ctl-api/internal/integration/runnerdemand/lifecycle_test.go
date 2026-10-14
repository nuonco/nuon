package runnerdemand_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	enumspb "go.temporal.io/api/enums/v1"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/demand"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/testworker"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

func testResume(t *testing.T, s testServices, namespace string) {
	f := newFixture(t, s, namespace, 2, time.Minute)
	f.start()
	for i := range f.jobs {
		f.awaitPark(i)
		var log app.LogStream
		require.NoError(t, s.DB.Where(app.LogStream{ID: f.log.ID}).First(&log).Error)
		require.True(t, log.Open)
		var processJobs int64
		require.NoError(t, s.DB.Model(&app.QueueSignal{}).Where(app.QueueSignal{OwnerID: f.jobs[i].ID, OwnerType: "runner_jobs"}).Count(&processJobs).Error)
		require.Zero(t, processJobs)
		execution := f.claim(i)
		f.request(http.MethodPost, f.jobURL(i)+"/executions", struct{}{}, http.StatusConflict, nil)
		var plan map[string]any
		f.request(http.MethodGet, f.jobURL(i)+"/plan", nil, http.StatusOK, &plan)
		require.Equal(t, "acme", plan["namespace"])
		result := f.complete(i, execution)
		var duplicate app.RunnerJobExecutionResult
		f.request(http.MethodPost, f.jobURL(i)+"/executions/"+execution.ID+"/result", map[string]any{"success": true, "contents": "duplicate"}, http.StatusCreated, &duplicate)
		require.Equal(t, result.ID, duplicate.ID)
		require.Equal(t, "completed by simulated runner", duplicate.Contents)
	}
	f.awaitWorkflow(app.StatusSuccess)
	var steps []app.WorkflowStep
	require.NoError(t, s.DB.Where(app.WorkflowStep{InstallWorkflowID: f.workflow.ID}).Find(&steps).Error)
	require.Len(t, steps, 1)
	require.Equal(t, f.step.ID, steps[0].ID)
	require.Equal(t, app.StatusSuccess, steps[0].Status.Status)
	var jobs []app.RunnerJob
	require.NoError(t, s.DB.Scopes(scopes.WithDisableViews).Where(app.RunnerJob{OwnerID: f.deploy.ID, OwnerType: "install_deploys"}).Order("created_at").Find(&jobs).Error)
	require.Len(t, jobs, 2)
	for i, j := range jobs {
		require.Equal(t, f.jobs[i].ID, j.ID)
		require.Equal(t, app.RunnerJobStatusFinished, j.Status)
	}
	var runs []app.WorkflowRun
	require.Eventually(t, func() bool {
		runs = nil
		return s.DB.Where(app.WorkflowRun{WorkflowID: f.workflow.ID}).Order("created_at").Find(&runs).Error == nil && len(runs) == 3 && runs[2].Status.Status == app.StatusSuccess
	}, time.Minute, 100*time.Millisecond)
	require.Len(t, runs, 3)
	require.Equal(t, app.WorkflowRunTypeResume, runs[2].Type)
	require.Equal(t, app.StatusSuccess, runs[2].Status.Status)
}

func testFastCompletion(t *testing.T, s testServices, namespace string) {
	f := newFixture(t, s, namespace, 1, time.Minute)
	require.NoError(t, s.DB.Model(f.step).Update("queue_signal", &signaldb.SignalData{Signal: &preparedJobsSignal{JobIDs: []string{f.jobs[0].ID}, HoldBeforePark: true}}).Error)
	f.start()
	f.awaitAvailable(0)
	f.complete(0, f.claim(0))
	require.Eventually(t, func() bool {
		var j app.RunnerJob
		return s.DB.Scopes(scopes.WithDisableViews).Where(app.RunnerJob{ID: f.jobs[0].ID}).First(&j).Error == nil && j.Status == app.RunnerJobStatusFinished
	}, 2*time.Minute, 100*time.Millisecond, "control plane must finish the job before its parent is allowed to park")
	var domainSignal app.QueueSignal
	for _, row := range f.signals() {
		if row.Type != preparedJobsSignalType {
			continue
		}
		domainSignal = row
	}
	require.NotEmpty(t, domainSignal.ID)
	resp, err := s.Temporal.DescribeWorkflowExecutionInNamespace(f.ctx, namespace, domainSignal.Workflow.ID, "")
	require.NoError(t, err)
	require.Equal(t, enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING, resp.GetWorkflowExecutionInfo().GetStatus(), "runner job completed while its parent is still unwinding")
	require.NoError(t, s.Temporal.SignalWorkflowInNamespace(f.ctx, namespace, domainSignal.Workflow.ID, "", "allow-park", true))
	f.awaitWorkflow(app.StatusSuccess)
}

func testParallelRetryAndRunner(t *testing.T, s testServices, namespace string) {
	f := newFixture(t, s, namespace, 1, time.Minute)
	require.NoError(t, s.DB.Model(f.group).Update("parallel", true).Error)
	retryStep := &app.WorkflowStep{InstallWorkflowID: f.workflow.ID, OwnerID: f.workflow.OwnerID, OwnerType: "installs", WorkflowStepGroupID: f.group.ID,
		Name: "needs manual retry", Idx: 50, Retryable: true, ExecutionType: app.WorkflowStepExecutionTypeSystem,
		Status: app.NewCompositeStatus(f.ctx, app.StatusPending), QueueSignal: &signaldb.SignalData{Signal: &testworker.ManualRetrySignal{}}}
	f.create(retryStep)
	downstreamGroup := s.Seeder.CreateWorkflowStepGroup(f.ctx, t, f.workflow.ID)
	require.NoError(t, s.DB.Model(downstreamGroup).Update("group_idx", 1).Error)
	downstream := &app.WorkflowStep{InstallWorkflowID: f.workflow.ID, OwnerID: f.workflow.OwnerID, OwnerType: "installs", WorkflowStepGroupID: downstreamGroup.ID,
		Name: "after both steps", Idx: 300, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
		Status: app.NewCompositeStatus(f.ctx, app.StatusPending), QueueSignal: &signaldb.SignalData{Signal: &preparedJobsSignal{}}}
	f.create(downstream)
	_, err := s.Queues.EnqueueSignal(f.ctx, &queueclient.EnqueueSignalRequest{QueueID: f.queue.ID, OwnerID: f.workflow.ID, OwnerType: "install_workflows", Signal: &executeflow.Signal{
		WorkflowID: f.workflow.ID, OwnerID: f.workflow.OwnerID, OwnerType: f.workflow.OwnerType,
		StepGroupQueueName: queuenames.InstallWorkflowStepGroupsQueueName, StepQueueName: queuenames.InstallWorkflowStepsQueueName,
		StepTargetQueueName: queuenames.InstallSignalsQueueName, GenerateStepsQueueName: queuenames.InstallGenerateStepsQueueName,
		Resident: true, ResidentIdleTimeout: 10 * time.Minute,
	}})
	require.NoError(t, err)
	f.awaitPark(0)
	f.complete(0, f.claim(0))
	require.Eventually(t, func() bool {
		var step app.WorkflowStep
		return s.DB.Where(app.WorkflowStep{ID: f.step.ID}).First(&step).Error == nil && step.Status.Status == app.StatusSuccess
	}, time.Minute, 100*time.Millisecond, "runner step must finalize without waiting for the resident idle timer or a manual retry")
	f.awaitWorkflow(app.StatusFailedPendingRetry)
	var failed app.WorkflowStep
	require.NoError(t, s.DB.Where(app.WorkflowStep{ID: retryStep.ID}).First(&failed).Error)
	require.Equal(t, app.StatusError, failed.Status.Status)
	require.Equal(t, string(directive.StepAwaitRetry), failed.ResultDirective)
	var next app.WorkflowStep
	require.NoError(t, s.DB.Where(app.WorkflowStep{ID: downstream.ID}).First(&next).Error)
	require.Equal(t, app.StatusPending, next.Status.Status)
	var beforeRetry app.WorkflowStep
	require.NoError(t, s.DB.Where(app.WorkflowStep{ID: f.step.ID}).First(&beforeRetry).Error)
	_, err = s.Flows.RetryStep(f.ctx, &flowclient.RetryStepRequest{InstallWorkflowID: f.workflow.ID, StepID: retryStep.ID})
	require.NoError(t, err)
	f.awaitWorkflow(app.StatusSuccess)
	var executions int64
	require.NoError(t, s.DB.Model(&app.RunnerJobExecution{}).Where(app.RunnerJobExecution{RunnerJobID: f.jobs[0].ID}).Count(&executions).Error)
	require.EqualValues(t, 1, executions)
	var completed app.WorkflowStep
	require.NoError(t, s.DB.Where(app.WorkflowStep{ID: f.step.ID}).First(&completed).Error)
	require.Equal(t, app.StatusSuccess, completed.Status.Status)
	require.Equal(t, beforeRetry.FinishedAt, completed.FinishedAt, "manual retry must not rerun the successful runner step")
	require.NoError(t, s.DB.Where(app.WorkflowStep{ID: retryStep.ID}).First(&failed).Error)
	require.Equal(t, app.StatusDiscarded, failed.Status.Status)
	require.True(t, failed.Retried)
	require.NoError(t, s.DB.Where(app.WorkflowStep{ID: downstream.ID}).First(&next).Error)
	require.Equal(t, app.StatusSuccess, next.Status.Status)
}

func testCancelOffline(t *testing.T, s testServices, namespace string) {
	f := newFixture(t, s, namespace, 1, time.Minute)
	f.start()
	f.awaitPark(0)
	var parkedSignal app.QueueSignal
	require.NoError(t, s.DB.Where(app.QueueSignal{OwnerID: f.step.ID, Type: preparedJobsSignalType}).First(&parkedSignal).Error)
	completedDependency := s.Seeder.CreateInstallDeploy(f.ctx, t, f.deploy.InstallComponentID, f.deploy.ComponentBuildID)
	require.NoError(t, s.DB.Scopes(scopes.WithDisableViews).Model(completedDependency).Updates(map[string]any{
		"status": app.InstallDeployStatusActive, "status_v2": app.NewCompositeStatus(f.ctx, app.Status(app.InstallDeployStatusActive)),
	}).Error)
	f.create(&app.RunnerJob{RunnerID: f.runner.ID, OwnerID: completedDependency.ID, OwnerType: "install_deploys", Status: app.RunnerJobStatusFinished,
		Type: app.RunnerJobTypeHelmChartDeploy, Group: app.RunnerJobGroupDeploy, Operation: app.RunnerJobOperationTypeApplyPlan,
		LogStreamID: &f.log.ID, Metadata: pgtype.Hstore{demand.SignalKey: &parkedSignal.ID, demand.StepKey: &f.step.ID, "flow_workflow_id": &f.workflow.ID}})
	var domainQueue app.Queue
	require.NoError(t, s.DB.Where(app.Queue{OwnerID: f.workflow.OwnerID, OwnerType: "installs", Name: queuenames.InstallSignalsQueueName}).First(&domainQueue).Error)
	targets := map[string]struct {
		deploy *app.InstallDeploy
		step   *app.WorkflowStep
		signal *app.QueueSignal
		job    *app.RunnerJob
	}{}
	for name, status := range map[string]struct {
		step   app.Status
		signal app.Status
		deploy app.InstallDeployStatus
		idx    int
	}{
		"completed":             {app.StatusSuccess, app.StatusSuccess, app.InstallDeployStatusActive, 50},
		"awaiting_finalization": {app.StatusWaitingForRunner, app.StatusWaitingForRunner, app.InstallDeployStatusExecuting, 200},
	} {
		deploy := s.Seeder.CreateInstallDeploy(f.ctx, t, f.deploy.InstallComponentID, f.deploy.ComponentBuildID)
		require.NoError(t, s.DB.Scopes(scopes.WithDisableViews).Model(deploy).Updates(map[string]any{"status": status.deploy, "status_v2": app.NewCompositeStatus(f.ctx, app.Status(status.deploy))}).Error)
		step := &app.WorkflowStep{InstallWorkflowID: f.workflow.ID, OwnerID: f.workflow.OwnerID, OwnerType: "installs", WorkflowStepGroupID: f.group.ID,
			Name: name, Idx: status.idx, Status: app.NewCompositeStatus(f.ctx, status.step), ExecutionType: app.WorkflowStepExecutionTypeSystem,
			StepTargetID: deploy.ID, StepTargetType: string(app.WorkflowStepTargetTypeInstallDeploys)}
		f.create(step)
		qs := &app.QueueSignal{QueueID: domainQueue.ID, OwnerID: step.ID, OwnerType: "install_workflow_steps", Type: preparedJobsSignalType,
			Signal: signaldb.SignalData{Signal: &preparedJobsSignal{}}, RunnerDemand: true, Status: app.NewCompositeStatus(f.ctx, status.signal),
			Workflow: signaldb.WorkflowRef{Namespace: namespace, IDTemplate: "test-demand-history-%s"}}
		f.create(qs)
		job := &app.RunnerJob{RunnerID: f.runner.ID, OwnerID: deploy.ID, OwnerType: "install_deploys", Status: app.RunnerJobStatusFinished,
			Type: app.RunnerJobTypeHelmChartDeploy, Group: app.RunnerJobGroupDeploy, Operation: app.RunnerJobOperationTypeApplyPlan,
			LogStreamID: &f.log.ID,
			Metadata:    pgtype.Hstore{demand.SignalKey: &qs.ID, demand.StepKey: &step.ID, "flow_workflow_id": &f.workflow.ID}}
		f.create(job)
		require.NoError(t, s.DB.Model(step).Updates(map[string]any{"runner_job_id": job.ID, "runner_job_signal_id": qs.ID}).Error)
		targets[name] = struct {
			deploy *app.InstallDeploy
			step   *app.WorkflowStep
			signal *app.QueueSignal
			job    *app.RunnerJob
		}{deploy, step, qs, job}
	}
	_, err := s.Flows.CancelWorkflow(f.ctx, &flowclient.CancelWorkflowRequest{InstallWorkflowID: f.workflow.ID})
	require.NoError(t, err)
	f.awaitWorkflow(app.StatusCancelled)
	f.request(http.MethodPost, f.jobURL(0)+"/executions", struct{}{}, http.StatusConflict, nil)
	var count int64
	require.NoError(t, s.DB.Model(&app.RunnerJobExecution{}).Where(app.RunnerJobExecution{RunnerJobID: f.jobs[0].ID}).Count(&count).Error)
	require.Zero(t, count)
	var preserved app.InstallDeploy
	require.NoError(t, s.DB.Scopes(scopes.WithDisableViews).Where(app.InstallDeploy{ID: completedDependency.ID}).First(&preserved).Error)
	require.Equal(t, app.InstallDeployStatusActive, preserved.Status, "completed dependency inside the unfinished signal must stay active")
	require.Equal(t, app.Status(app.InstallDeployStatusActive), preserved.StatusV2.Status)
	for name, target := range targets {
		var deploy app.InstallDeploy
		require.NoError(t, s.DB.Scopes(scopes.WithDisableViews).Where(app.InstallDeploy{ID: target.deploy.ID}).First(&deploy).Error)
		var step app.WorkflowStep
		require.NoError(t, s.DB.Where(app.WorkflowStep{ID: target.step.ID}).First(&step).Error)
		var qs app.QueueSignal
		require.NoError(t, s.DB.Where(app.QueueSignal{ID: target.signal.ID}).First(&qs).Error)
		var job app.RunnerJob
		require.NoError(t, s.DB.Scopes(scopes.WithDisableViews).Where(app.RunnerJob{ID: target.job.ID}).First(&job).Error)
		require.Equal(t, app.RunnerJobStatusFinished, job.Status, name)
		if name == "completed" {
			require.Equal(t, app.InstallDeployStatusActive, deploy.Status)
			require.Equal(t, app.Status(app.InstallDeployStatusActive), deploy.StatusV2.Status)
			require.Equal(t, app.StatusSuccess, step.Status.Status)
			require.Equal(t, app.StatusSuccess, qs.Status.Status)
		} else {
			require.Equal(t, app.InstallDeployStatusCancelled, deploy.Status)
			require.Equal(t, app.StatusCancelled, step.Status.Status)
			require.Equal(t, app.StatusCancelled, qs.Status.Status)
		}
	}
}

func testCancelExecuting(t *testing.T, s testServices, namespace string) {
	f := newFixture(t, s, namespace, 1, time.Minute)
	f.start()
	f.awaitPark(0)
	execution := f.claim(0)
	_, err := s.Flows.CancelWorkflow(f.ctx, &flowclient.CancelWorkflowRequest{InstallWorkflowID: f.workflow.ID})
	require.NoError(t, err)
	f.awaitWorkflow(app.StatusCancelled)
	url := f.jobURL(0) + "/executions/" + execution.ID
	f.request(http.MethodPost, url+"/result", map[string]any{"success": true}, http.StatusConflict, nil)
	f.request(http.MethodPatch, url, map[string]any{"status": "finished"}, http.StatusConflict, nil)
	var results int64
	require.NoError(t, s.DB.Model(&app.RunnerJobExecutionResult{}).Where(app.RunnerJobExecutionResult{RunnerJobExecutionID: execution.ID}).Count(&results).Error)
	require.Zero(t, results)
}

func testCancelJob(t *testing.T, s testServices, namespace string) {
	f := newFixture(t, s, namespace, 1, time.Minute)
	f.start()
	f.awaitPark(0)
	f.request(http.MethodPost, f.public.URL+"/v1/runner-jobs/"+f.jobs[0].ID+"/cancel", struct{}{}, http.StatusAccepted, nil)
	f.request(http.MethodPost, f.jobURL(0)+"/executions", struct{}{}, http.StatusConflict, nil)
	f.awaitWorkflow(app.StatusError)
}

func testExecutionTimeout(t *testing.T, s testServices, namespace string) {
	const timeout = 8 * time.Second
	f := newFixture(t, s, namespace, 1, timeout)
	f.start()
	f.awaitPark(0)
	select {
	case <-time.After(time.Until(f.jobs[0].CreatedAt.Add(timeout + time.Second))):
	case <-t.Context().Done():
		t.Fatal(t.Context().Err())
	}
	execution := f.claim(0)
	require.Greater(t, execution.CreatedAt.Sub(f.jobs[0].CreatedAt), timeout)
	var job app.RunnerJob
	require.NoError(t, s.DB.Scopes(scopes.WithDisableViews).Where(app.RunnerJob{ID: f.jobs[0].ID}).First(&job).Error)
	require.WithinDuration(t, execution.CreatedAt, job.StartedAt, time.Millisecond)
	f.request(http.MethodPatch, f.jobURL(0)+"/executions/"+execution.ID, map[string]any{"status": "in-progress"}, http.StatusOK, nil)
	require.Eventually(t, func() bool {
		var j app.RunnerJob
		return s.DB.Scopes(scopes.WithDisableViews).Where(app.RunnerJob{ID: f.jobs[0].ID}).First(&j).Error == nil && j.Status == app.RunnerJobStatusTimedOut
	}, time.Minute, 100*time.Millisecond)
	require.GreaterOrEqual(t, time.Since(execution.CreatedAt), timeout)
	f.request(http.MethodPost, f.jobURL(0)+"/executions/"+execution.ID+"/result", map[string]any{"success": true}, http.StatusConflict, nil)
	f.awaitWorkflow(app.StatusError)
}
