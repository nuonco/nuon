package jobloop

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon-runner-go/models"
	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"go.uber.org/zap"
)

func (h *jobLoop) execActionSandboxStep(ctx context.Context, job *models.AppRunnerJob) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	// fetch the plan json
	l.Info("fetching actions job plan")
	planJSON, err := h.apiClient.GetJobPlanJSON(ctx, job.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get job plan")
	}

	// parse the plan
	var plan plantypes.ActionWorkflowRunPlan
	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		return errors.Wrap(err, "unable to parse action workflow run plan")
	}

	// fetch the run object
	run, err := h.apiClient.GetInstallActionWorkflowRun(ctx,
		plan.InstallID,
		plan.ID,
	)
	if err != nil {
		return errors.Wrap(err, "unable to get action workflow run")
	}

	// fetch the workflow config
	l.Info("fetching actions workflow config")
	cfg, err := h.apiClient.GetActionWorkflowConfig(ctx, run.ActionWorkflowConfigID)
	if err != nil {
		return errors.Wrap(err, "unable to get action workflow run")
	}

	for idx, step := range run.Steps {
		stepCfg := cfg.Steps[idx]
		l = l.With(
			zap.String("workflow_step_name", stepCfg.Name),
			zap.String("step_run_id", step.ID),
		)

		l.Info(fmt.Sprintf("executing step %s (%d of %d)", stepCfg.Name, idx+1, len(run.Steps)))

		_, err := h.apiClient.UpdateInstallActionWorkflowRunStep(ctx, plan.InstallID, cfg.ActionWorkflowID, step.ID, &models.ServiceUpdateInstallActionWorkflowRunStepRequest{
			Status:            models.AppInstallActionWorkflowRunStepStatusFinished,
			ExecutionDuration: int64(time.Second * 5),
		})
		if err != nil {
			return errors.Wrap(err, "unable to update step status")
		}
	}

	return nil
}
