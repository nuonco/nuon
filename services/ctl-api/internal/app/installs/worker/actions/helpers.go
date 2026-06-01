package actions

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/state/statepartialgenerate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

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
	planResponse, err := plan.AwaitCreateActionWorkflowRunPlan(ctx, &plan.CreateActionRunPlanRequest{
		ActionWorkflowRunID: actionWorkflowRunID,
		WorkflowID:          fmt.Sprintf("%s-create-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to create plan")
		return errors.Wrap(err, "unable to create plan")
	}

	l.Info("creating runner job to execute action")
	metadata := map[string]string{
		"install_id":             installID,
		"action_workflow_run_id": run.ID,
	}

	if run.ActionWorkflowConfigID.Valid {
		metadata["action_workflow_name"] = run.ActionWorkflowConfig.ActionWorkflow.Name
		metadata["action_workflow_id"] = run.ActionWorkflowConfig.ActionWorkflowID
	} else {
		actionName := "Adhoc action"
		if len(run.Steps) > 0 && run.Steps[0].AdHocConfig != nil && run.Steps[0].AdHocConfig.Name != "" {
			actionName = run.Steps[0].AdHocConfig.Name
		}
		metadata["action_workflow_name"] = actionName
		metadata["action_workflow_id"] = ""
	}

	runnerJob, err := activities.AwaitCreateActionWorkflowRunRunnerJob(ctx, &activities.CreateActionWorkflowRunRunnerJob{
		ActionWorkflowRunID: actionWorkflowRunID,
		RunnerID:            run.Install.RunnerID,
		LogStreamID:         ls.ID,
		Metadata:            metadata,
	})
	if err != nil {
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to create job")
		return errors.Wrap(err, "unable to create runner job")
	}

	planJSON, err := json.Marshal(planResponse.Plan)
	if err != nil {
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to create job")
		return errors.Wrap(err, "unable to convert plan to json")
	}

	compositePlan := plantypes.CompositePlan{
		ActionWorkflowRunPlan: planResponse.Plan,
	}

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:         runnerJob.ID,
		PlanJSON:      string(planJSON),
		CompositePlan: compositePlan,
	}); err != nil {
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to save job plan")
		return errors.Wrap(err, "unable to save runner job plan")
	}

	if err := activities.AwaitRecordInstallRoleUsage(ctx, &activities.RecordInstallRoleUsageRequest{
		InstallID:     installID,
		RunnerJobID:   runnerJob.ID,
		RoleSelection: planResponse.RoleSelection,
	}); err != nil {
		w.updateActionRunStatus(ctx, run.ID, app.InstallActionRunStatusError, "unable to record install role usage")
		return errors.Wrap(err, "unable to record install role usage")
	}

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

	sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   installID,
		OwnerType: "installs",
		QueueName: installshelpers.InstallStateManagerQueueName,
		Signal: &statepartialgenerate.Signal{
			InstallID:       installID,
			Targets:         statemanager.TargetsForHint(statemanager.HintActionRan, ""),
			ForceAll:        true,
			TriggeredByID:   actionWorkflowRunID,
			TriggeredByType: "install_action_workflow_runs",
		},
	})

	return nil
}
