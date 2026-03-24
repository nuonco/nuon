package worker

import (
	"fmt"

	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

func cronShutdownVMWorkflowID(runnerID string) string {
	return fmt.Sprintf("cron-shutdown-vm-%s", runnerID)
}

func (w *Workflows) startCronShutdownVMWorkflow(ctx workflow.Context, req CronShutdownVMRequest) {
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:            cronShutdownVMWorkflowID(req.RunnerID),
		CronSchedule:          "0 12 * * 0", // 4am PST (5am PDT) every Sunday
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		ParentClosePolicy:     enumsv1.PARENT_CLOSE_POLICY_TERMINATE,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)
	workflow.ExecuteChildWorkflow(ctx, w.CronShutdownVM, &req)
}

type CronShutdownVMRequest struct {
	RunnerID string `validate:"required" json:"runner_id"`
}

// @temporal-gen-v2 workflow
// @execution-timeout 3m
// @task-timeout 5m
func (w *Workflows) CronShutdownVM(ctx workflow.Context, req *CronShutdownVMRequest) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get logger")
	}
	l.Info("cron shutdown vm: starting",
		zap.String("runner_id", req.RunnerID),
	)

	runner, err := activities.AwaitGetByRunnerID(ctx, req.RunnerID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner")
	}
	if runner.RunnerGroup.Type != app.RunnerGroupTypeInstall {
		l.Info("cron shutdown vm: skipping non-install runner",
			zap.String("runner_id", req.RunnerID),
			zap.String("group_type", string(runner.RunnerGroup.Type)),
		)
		return nil
	}

	l.Info("cron shutdown vm: creating mng vm shutdown job",
		zap.String("runner_id", req.RunnerID),
	)

	runnerJob, err := w.createMngJob(ctx, req.RunnerID, app.RunnerJobTypeMngVMShutDown, map[string]string{
		"shutdown_type": "vm",
	})
	if err != nil {
		return errors.Wrap(err, "unable to create vm shutdown job")
	}

	if err := activities.AwaitUpdateJobStatus(ctx, activities.UpdateJobStatusRequest{
		JobID:             runnerJob.ID,
		Status:            app.RunnerJobStatusAvailable,
		StatusDescription: string(app.RunnerJobStatusAvailable),
	}); err != nil {
		return errors.Wrap(err, "unable to mark vm shutdown job available")
	}

	l.Info("cron shutdown vm: job dispatched",
		zap.String("runner_id", req.RunnerID),
		zap.String("job_id", runnerJob.ID),
	)

	return nil
}
