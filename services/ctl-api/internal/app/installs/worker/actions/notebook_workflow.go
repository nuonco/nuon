package actions

import (
	"encoding/json"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

// NotebookWorkflowType is the registered Temporal workflow type name.
const NotebookWorkflowType = "NotebookWorkflow"

// NotebookRunCellUpdateName is the Temporal update handler name used to run a
// cell on a warm notebook workflow.
const NotebookRunCellUpdateName = "run-cell"

// notebookIdleTimeout is how long the warm workflow waits with no pending runs
// before completing. The next RunCell re-starts it via update-with-start.
const notebookIdleTimeout = 15 * time.Minute

// NotebookWorkflowID returns the deterministic per-notebook workflow ID.
func NotebookWorkflowID(notebookID string) string {
	return "notebook-" + notebookID
}

type NotebookWorkflowRequest struct {
	NotebookID string
	// State carries pending runs across continue-as-new. Nil on first start.
	State *NotebookWorkflowState
}

type NotebookWorkflowState struct {
	Pending []NotebookRunRef
}

// NotebookRunRef is everything the workflow needs to execute one cell run,
// carried explicitly so the long-lived workflow never relies on a stale start
// context for per-run ownership/audit.
type NotebookRunRef struct {
	NotebookCellRunID string
	ActionRunID       string
	OrgID             string
	InstallID         string
	TriggeredByID     string
	RunnerID          string
	Role              string
}

type RunCellUpdate struct {
	CellID         string
	IdempotencyKey string
	OrgID          string
	InstallID      string
	TriggeredByID  string
}

type RunCellUpdateResult struct {
	NotebookCellRunID          string
	InstallActionWorkflowRunID string
}

// NotebookWorkflow is a warm, long-lived per-notebook workflow. Running a cell
// is a Temporal update ("run-cell") that creates the run rows and enqueues the
// run, returning IDs immediately; the workflow's main loop then dispatches the
// runner job. This skips the cold nested install-workflow step tree that a
// normal adhoc action run pays on every invocation.
func (w *Workflows) NotebookWorkflow(ctx workflow.Context, req NotebookWorkflowRequest) error {
	l := workflow.GetLogger(ctx)
	state := req.State
	if state == nil {
		state = &NotebookWorkflowState{}
	}

	if err := workflow.SetUpdateHandlerWithOptions(ctx, NotebookRunCellUpdateName,
		func(ctx workflow.Context, ureq RunCellUpdate) (RunCellUpdateResult, error) {
			resp, err := activities.AwaitCreateNotebookCellRun(ctx, &activities.CreateNotebookCellRunRequest{
				NotebookID:     req.NotebookID,
				CellID:         ureq.CellID,
				IdempotencyKey: ureq.IdempotencyKey,
				OrgID:          ureq.OrgID,
				InstallID:      ureq.InstallID,
				TriggeredByID:  ureq.TriggeredByID,
			})
			if err != nil {
				return RunCellUpdateResult{}, err
			}

			// Only enqueue genuinely new runs; an idempotency-key hit means the
			// run was already created (and enqueued) by a prior request.
			if !resp.AlreadyDispatched {
				state.Pending = append(state.Pending, NotebookRunRef{
					NotebookCellRunID: resp.NotebookCellRunID,
					ActionRunID:       resp.InstallActionWorkflowRunID,
					OrgID:             ureq.OrgID,
					InstallID:         ureq.InstallID,
					TriggeredByID:     ureq.TriggeredByID,
					RunnerID:          resp.RunnerID,
					Role:              resp.Role,
				})
			}

			return RunCellUpdateResult{
				NotebookCellRunID:          resp.NotebookCellRunID,
				InstallActionWorkflowRunID: resp.InstallActionWorkflowRunID,
			}, nil
		}, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}

	for {
		if len(state.Pending) == 0 {
			ok, err := workflow.AwaitWithTimeout(ctx, notebookIdleTimeout, func() bool {
				return len(state.Pending) > 0
			})
			if err != nil {
				return err
			}
			// Idle with no in-flight updates: let the workflow complete. The
			// next RunCell re-starts it via update-with-start.
			if !ok && workflow.AllHandlersFinished(ctx) {
				return nil
			}
			if !ok {
				continue
			}
		}

		// Continue-as-new only between runs, once all update handlers have
		// finished. Pending runs are carried forward.
		if workflow.GetInfo(ctx).GetContinueAsNewSuggested() && workflow.AllHandlersFinished(ctx) {
			ctx = cctx.SetLogStreamWorkflowContext(ctx, nil)
			return workflow.NewContinueAsNewError(ctx, w.NotebookWorkflow, NotebookWorkflowRequest{
				NotebookID: req.NotebookID,
				State:      state,
			})
		}

		ref := state.Pending[0]
		state.Pending = state.Pending[1:]
		// A single bad run must not stop the notebook; executeNotebookRun
		// records terminal status on the run and returns.
		w.executeNotebookRun(ctx, ref)
		l.Info("notebook cell run finished", zap.String("action_run_id", ref.ActionRunID))
	}
}

// executeNotebookRun dispatches one cell run: log stream, plan, runner job, and
// job execution. It mirrors the adhoc action-run execution path but scopes
// child-workflow IDs to the action run (the parent workflow ID is stable across
// many runs) and mirrors status onto the NotebookCellRun for the UI.
func (w *Workflows) executeNotebookRun(ctx workflow.Context, ref NotebookRunRef) {
	l := workflow.GetLogger(ctx)

	// Per-run context: the long-lived workflow serves many accounts, so set
	// org/account explicitly and start each run with a clean log stream.
	runCtx := cctx.SetOrgIDWorkflowContext(ctx, ref.OrgID)
	runCtx = cctx.SetAccountIDWorkflowContext(runCtx, ref.TriggeredByID)
	runCtx = cctx.SetLogStreamWorkflowContext(runCtx, nil)

	fail := func(msg string) {
		w.updateNotebookRunStatus(runCtx, ref, app.InstallActionRunStatusError, msg)
	}

	w.updateNotebookRunStatus(runCtx, ref, app.InstallActionRunStatusInProgress, "in-progress")

	ls, err := activities.AwaitCreateLogStream(runCtx, activities.CreateLogStreamRequest{
		ActionWorkflowRunID: ref.ActionRunID,
	})
	if err != nil {
		fail("unable to create log stream")
		return
	}
	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(runCtx, ls.ID)
	}()
	runCtx = cctx.SetLogStreamWorkflowContext(runCtx, ls)

	// Surface the log stream on the cell run ASAP so the UI can start tailing.
	if err := activities.AwaitUpdateNotebookCellRun(runCtx, &activities.UpdateNotebookCellRunRequest{
		NotebookCellRunID: ref.NotebookCellRunID,
		LogStreamID:       ls.ID,
	}); err != nil {
		l.Error("unable to record log stream on cell run", zap.Error(err))
	}

	planResponse, err := plan.AwaitCreateActionWorkflowRunPlan(runCtx, &plan.CreateActionRunPlanRequest{
		ActionWorkflowRunID: ref.ActionRunID,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("notebook-plan-%s", ref.ActionRunID),
	})
	if err != nil {
		fail("unable to create plan")
		return
	}

	runnerJob, err := activities.AwaitCreateActionWorkflowRunRunnerJob(runCtx, &activities.CreateActionWorkflowRunRunnerJob{
		ActionWorkflowRunID: ref.ActionRunID,
		RunnerID:            ref.RunnerID,
		LogStreamID:         ls.ID,
		Metadata: map[string]string{
			"install_id":             ref.InstallID,
			"action_workflow_run_id": ref.ActionRunID,
			"notebook_cell_run_id":   ref.NotebookCellRunID,
		},
	})
	if err != nil {
		fail("unable to create job")
		return
	}

	if err := activities.AwaitUpdateNotebookCellRun(runCtx, &activities.UpdateNotebookCellRunRequest{
		NotebookCellRunID: ref.NotebookCellRunID,
		RunnerJobID:       runnerJob.ID,
	}); err != nil {
		l.Error("unable to record runner job on cell run", zap.Error(err))
	}

	planJSON, err := json.Marshal(planResponse.Plan)
	if err != nil {
		fail("unable to create job")
		return
	}
	if err := activities.AwaitSaveRunnerJobPlan(runCtx, &activities.SaveRunnerJobPlanRequest{
		JobID:         runnerJob.ID,
		PlanJSON:      string(planJSON),
		CompositePlan: plantypes.CompositePlan{ActionWorkflowRunPlan: planResponse.Plan},
	}); err != nil {
		fail("unable to save job plan")
		return
	}

	if err := activities.AwaitRecordInstallRoleUsage(runCtx, &activities.RecordInstallRoleUsageRequest{
		InstallID:     ref.InstallID,
		RunnerJobID:   runnerJob.ID,
		RoleSelection: planResponse.RoleSelection,
	}); err != nil {
		fail("unable to record install role usage")
		return
	}

	if _, err := job.AwaitExecuteJob(runCtx, &job.ExecuteJobRequest{
		RunnerID: ref.RunnerID,
		JobID:    runnerJob.ID,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("notebook-exec-job-%s", ref.ActionRunID),
	}); err != nil {
		fail(job.JobErrorMessage(err, "notebook cell job failed"))
		return
	}

	w.updateNotebookRunStatus(runCtx, ref, app.InstallActionRunStatusFinished, "finished")
}

// updateNotebookRunStatus mirrors a status onto the underlying
// InstallActionWorkflowRun (v1 + v2) and the NotebookCellRun row.
func (w *Workflows) updateNotebookRunStatus(ctx workflow.Context, ref NotebookRunRef, status app.InstallActionWorkflowRunStatus, msg string) {
	l := workflow.GetLogger(ctx)

	if err := activities.AwaitUpdateInstallWorkflowRunStatus(ctx, activities.UpdateInstallWorkflowRunStatusRequest{
		RunID:             ref.ActionRunID,
		Status:            status,
		StatusDescription: msg,
	}); err != nil {
		l.Error("unable to update action run status", zap.String("run-id", ref.ActionRunID), zap.Error(err))
	}

	if err := statusactivities.AwaitUpdateInstallWorkflowRunStatusV2(ctx, statusactivities.UpdateInstallWorkflowRunStatusV2Request{
		RunID:             ref.ActionRunID,
		Status:            status,
		StatusDescription: msg,
	}); err != nil {
		l.Error("unable to update action run status v2", zap.String("run-id", ref.ActionRunID), zap.Error(err))
	}

	if err := activities.AwaitUpdateNotebookCellRun(ctx, &activities.UpdateNotebookCellRunRequest{
		NotebookCellRunID: ref.NotebookCellRunID,
		Status:            status,
		StatusDescription: msg,
	}); err != nil {
		l.Error("unable to update notebook cell run status", zap.String("cell-run-id", ref.NotebookCellRunID), zap.Error(err))
	}
}
