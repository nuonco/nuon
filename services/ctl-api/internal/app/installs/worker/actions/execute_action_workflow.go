package actions

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/job"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) ExecuteActionWorkflow(ctx workflow.Context, req signals.RequestSignal) error {
	installActionWorkflow, err := activities.AwaitGetInstallActionWorkflowByID(ctx, req.InstallActionWorkflowTrigger.InstallActionWorkflowID)
	if err != nil {
		return errors.Wrap(err, "unable to get install action workflow")
	}

	actionWorkflowRun, err := activities.AwaitCreateActionWorkflowRun(ctx, &activities.CreateActionWorkflowRunRequest{
		InstallActionWorkflowID: installActionWorkflow.ID,
		ActionWorkflowID:        installActionWorkflow.ActionWorkflowID,
		InstallID:               installActionWorkflow.InstallID,
		InstallWorkflowID:       req.InstallWorkflowID,
		TriggerType:             req.InstallActionWorkflowTrigger.TriggerType,
		TriggeredByID:           req.InstallActionWorkflowTrigger.TriggeredByID,
		TriggeredByType:         req.InstallActionWorkflowTrigger.TriggeredByType,
		RunEnvVars:              generics.ToPtrStringMap(req.InstallActionWorkflowTrigger.RunEnvVars),
	})
	if err != nil {
		return errors.Wrap(err, "unable to create action workflow run")
	}

	if req.WorkflowStepID != "" {
		if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
			StepID:         req.WorkflowStepID,
			StepTargetID:   actionWorkflowRun.ID,
			StepTargetType: plugins.TableName(w.db, actionWorkflowRun),
		}); err != nil {
			return errors.Wrap(err, "unable to update install action workflow")
		}
	}

	if err := w.executeActionWorkflowRun(ctx, installActionWorkflow.InstallID, actionWorkflowRun.ID); err != nil {
		return errors.Wrap(err, "unable to create action workflow run")
	}

	return nil
}

func (w *Workflows) executeActionWorkflowRun(ctx workflow.Context, installID, actionWorkflowRunID string) error {
	run, err := activities.AwaitGetInstallActionWorkflowRunByRunID(ctx, actionWorkflowRunID)
	if err != nil {
		return errors.Wrap(err, "unable to get action workflow run")
	}

	l, err := log.WorkflowLogger(ctx)
	if err == nil {
		l.Warn("creating a new logger for executing action")
	}
	parentLS, _ := cctx.GetLogStreamWorkflow(ctx)

	lsReq := activities.CreateLogStreamRequest{
		ActionWorkflowRunID: actionWorkflowRunID,
	}
	if parentLS != nil {
		lsReq.ParentLogStreamID = parentLS.ID
	}
	w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusInProgress, "in-progress")
	ls, err := activities.AwaitCreateLogStream(ctx, lsReq)
	if err != nil {
		return errors.Wrap(err, "unable to create log stream")
	}

	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, ls.ID)
	}()
	ctx = cctx.SetLogStreamWorkflowContext(ctx, ls)

	l, err = log.WorkflowLogger(ctx)
	if err != nil {
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to create log stream")
		return errors.Wrap(err, "unable to set log stream on context")
	}

	ls, err = cctx.GetLogStreamWorkflow(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get log stream")
	}

	l.Info("creating plan for executing action run")
	runPlan, err := plan.AwaitCreateActionWorkflowRunPlan(ctx, &plan.CreateActionRunPlanRequest{
		ActionWorkflowRunID: actionWorkflowRunID,
		WorkflowID:          fmt.Sprintf("%s-create-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to create plan")
		return errors.Wrap(err, "unable to create plan")
	}

	// execute job
	l.Info("creating runner job to execute action")
	runnerJob, err := activities.AwaitCreateActionWorkflowRunRunnerJob(ctx, &activities.CreateActionWorkflowRunRunnerJob{
		ActionWorkflowRunID: actionWorkflowRunID,
		RunnerID:            run.Install.RunnerID,
		LogStreamID:         ls.ID,
		Metadata: map[string]string{
			"install_id":             installID,
			"action_workflow_name":   run.ActionWorkflowConfig.ActionWorkflow.Name,
			"action_workflow_run_id": run.ID,
			"action_workflow_id":     run.ActionWorkflowConfig.ActionWorkflowID,
		},
	})
	if err != nil {
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to create job")
		return errors.Wrap(err, "unable to create runner job")
	}

	// save runner job plan
	planJSON, err := json.Marshal(runPlan)
	if err != nil {
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to create job")
		return errors.Wrap(err, "unable to convert plan to json")
	}

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:         runnerJob.ID,
		PlanJSON:      string(planJSON),
		CompositePlan: plantypes.CompositePlan{ActionWorkflowRunPlan: runPlan},
	}); err != nil {
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to save job plan")
		return errors.Wrap(err, "unable to save runner job plan")
	}

	planJSON = nil
	runPlan = nil

	// now queue and execute the job
	l.Info("executing runner job")
	_, err = job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		RunnerID:   run.Install.RunnerID,
		JobID:      runnerJob.ID,
		WorkflowID: "actions-install-run-exec-job" + run.ID,
	})
	if err != nil {
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "job failed")
		return errors.Wrap(err, "runner job failed")
	}

	w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusFinished, "finished")

	_, err = state.AwaitGenerateState(ctx, &state.GenerateStateRequest{
		InstallID:       installID,
		TriggeredByID:   actionWorkflowRunID,
		TriggeredByType: plugins.TableName(w.db, run),
	})
	if err != nil {
		return errors.Wrap(err, "unable to generate state")
	}

	return nil
}
