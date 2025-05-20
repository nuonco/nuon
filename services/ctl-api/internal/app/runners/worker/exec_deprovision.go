package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/workflows/types/executors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

func (w *Workflows) executeDeprovisionOrgRunner(ctx workflow.Context, runnerID string, sandboxMode bool) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	runner, err := activities.AwaitGet(ctx, activities.GetRequest{
		RunnerID: runnerID,
	})
	if err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to get runner")
		return fmt.Errorf("unable to get runner: %w", err)
	}

	if runner.RunnerGroup.Platform == app.AppRunnerTypeLocal {
		l.Info("skipping local runner")
		return nil
	}
	if runner.Org.OrgType == app.OrgTypeIntegration {
		return nil
	}

	req := &executors.DeprovisionRunnerRequest{
		RunnerID: runnerID,
	}
	var resp executors.DeprovisionRunnerResponse
	err = w.execChildWorkflow(ctx, runnerID, executors.DeprovisionRunnerWorkflowName, sandboxMode, req, &resp)
	if err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to deprovision runner")
		return fmt.Errorf("unable to deprovision runner: %w", err)
	}

	return nil
}
