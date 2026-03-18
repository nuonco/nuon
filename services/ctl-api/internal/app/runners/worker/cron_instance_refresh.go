package worker

import (
	"fmt"

	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

func instanceRefreshWorkflowID(runnerID string) string {
	return fmt.Sprintf("instance-refresh-%s", runnerID)
}

func (w *Workflows) startInstanceRefreshWorkflow(ctx workflow.Context, req InstanceRefreshRequest) {
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:            instanceRefreshWorkflowID(req.RunnerID),
		CronSchedule:          "*/5 * * * *", // 3am PST (4am PDT)
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		ParentClosePolicy:     enumsv1.PARENT_CLOSE_POLICY_TERMINATE,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)
	workflow.ExecuteChildWorkflow(ctx, w.InstanceRefresh, &req)
}

type InstanceRefreshRequest struct {
	RunnerID string `validate:"required" json:"runner_id"`
}

// @temporal-gen-v2 workflow
// @execution-timeout 3m
// @task-timeout 5m
func (w *Workflows) InstanceRefresh(ctx workflow.Context, req *InstanceRefreshRequest) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get logger")
	}

	l.Info("executing instance refresh",
		zap.String("runner_id", req.RunnerID),
	)

	runner, err := activities.AwaitGetByRunnerID(ctx, req.RunnerID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner")
	}

	if runner.RunnerGroup.Type != app.RunnerGroupTypeInstall {
		l.Info("skipping instance refresh for non-install runner",
			zap.String("runner_id", req.RunnerID),
			zap.String("group_type", string(runner.RunnerGroup.Type)),
		)
		return nil
	}

	noop := generics.SliceContains(runner.Status, []app.RunnerStatus{
		app.RunnerStatusPending,
		app.RunnerStatusProvisioning,
		app.RunnerStatusDeprovisioning,
		app.RunnerStatusReprovisioning,
		app.RunnerStatusDeprovisioned,
		app.RunnerStatusOffline,
		app.RunnerStatusAwaitingInstallStackRun,
	})
	if noop {
		l.Info("skipping instance refresh, runner not in operational state",
			zap.String("runner_id", req.RunnerID),
			zap.String("status", string(runner.Status)),
		)
		return nil
	}

	runnerJob, err := w.createRunnerVMShutDownJob(ctx, req.RunnerID, map[string]string{
		"shutdown_type": "vm",
	})
	if err != nil {
		return errors.Wrap(err, "unable to create instance refresh job")
	}

	if err := activities.AwaitUpdateJobStatus(ctx, activities.UpdateJobStatusRequest{
		JobID:             runnerJob.ID,
		Status:            app.RunnerJobStatusAvailable,
		StatusDescription: string(app.RunnerJobStatusAvailable),
	}); err != nil {
		return errors.Wrap(err, "unable to mark instance refresh job available")
	}

	return nil
}
