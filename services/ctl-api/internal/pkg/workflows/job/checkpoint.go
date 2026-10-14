package job

import (
	"encoding/json"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/continuation"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job/activities"
)

func DemandEnabled(ctx workflow.Context) bool {
	return continuation.Scope(ctx) != "" && workflow.GetVersion(ctx, "runner-demand-checkpoint-v1", workflow.DefaultVersion, 1) != workflow.DefaultVersion
}

func LoadCheckpoint(ctx workflow.Context, phase string, state any) (bool, error) {
	if !DemandEnabled(ctx) {
		return false, nil
	}
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{StartToCloseTimeout: time.Minute})
	var data []byte
	err := workflow.ExecuteActivity(ctx, (*activities.Activities).LoadRunnerJobCheckpoint, continuation.Scope(ctx), phase).Get(ctx, &data)
	if err != nil || len(data) == 0 {
		return false, err
	}
	return true, json.Unmarshal(data, state)
}

func SaveCheckpoint(ctx workflow.Context, phase string, state any) error {
	if !DemandEnabled(ctx) {
		return nil
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{StartToCloseTimeout: time.Minute})
	return workflow.ExecuteActivity(ctx, (*activities.Activities).SaveRunnerJobCheckpoint, continuation.Scope(ctx), phase, data).Get(ctx, nil)
}

func IsSuspended(err error) bool { return callback.IsSuspended(err) }

func Execute(ctx workflow.Context, req *ExecuteJobRequest, opts *workflow.ChildWorkflowOptions) (app.RunnerJobStatus, error) {
	if !DemandEnabled(ctx) {
		return AwaitExecuteJob(ctx, req, opts)
	}
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{StartToCloseTimeout: time.Minute})
	var status app.RunnerJobStatus
	err := workflow.ExecuteActivity(ctx, (*activities.Activities).PublishRunnerJob, continuation.Scope(ctx), req.JobID).Get(ctx, &status)
	if err != nil {
		return status, err
	}
	if !status.IsTerminal() {
		return status, callback.Suspended()
	}
	if status != app.RunnerJobStatusFinished {
		return status, temporal.NewNonRetryableApplicationError("runner job did not succeed: "+string(status), "RUNNER_JOB_FAILED", nil)
	}
	return status, nil
}
