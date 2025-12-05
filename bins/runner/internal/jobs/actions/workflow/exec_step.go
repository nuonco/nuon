package workflow

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"

	"github.com/nuonco/nuon-runner-go/models"
)

func (h *handler) updateStepStatus(ctx context.Context, stepID string, startTS time.Time, status models.AppInstallActionWorkflowRunStepStatus) error {
	_, err := h.apiClient.UpdateInstallActionWorkflowRunStep(ctx, h.state.plan.InstallID, h.state.workflowCfg.ActionWorkflowID, stepID, &models.ServiceUpdateInstallActionWorkflowRunStepRequest{
		Status:            status,
		ExecutionDuration: time.Now().Sub(startTS).Nanoseconds(),
	})
	if err != nil {
		return errors.Wrap(err, "unable to update step status")
	}

	return nil
}

func (h *handler) executeWorkflowStep(ctx, execCtx context.Context, step *models.AppInstallActionWorkflowRunStep, cfg *models.AppActionWorkflowStepConfig, stepPlan *plantypes.ActionWorkflowRunStepPlan) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}
	startTS := time.Now()

	l = l.With(
		zap.String("workflow_step_name", cfg.Name),
		zap.String("step_run_id", step.ID),
	)

	if err := h.updateStepStatus(ctx, step.ID, startTS, models.AppInstallActionWorkflowRunStepStatusInDashProgress); err != nil {
		return errors.Wrap(err, "unable to update status")
	}

	if err := h.createExecEnv(execCtx, l, stepPlan.GitSource, cfg); err != nil {
		h.updateStepStatus(ctx, step.ID, startTS, models.AppInstallActionWorkflowRunStepStatusError)
		return errors.Wrap(err, "unable to create exec env")
	}

	if err := h.execCommand(execCtx, l, cfg, stepPlan.GitSource, stepPlan.InterpolatedEnvVars); err != nil {
		status := models.AppInstallActionWorkflowRunStepStatusError
		if errors.Is(err, context.DeadlineExceeded) {
			status = models.AppInstallActionWorkflowRunStepStatusTimedDashOut
		}

		h.updateStepStatus(ctx, step.ID, startTS, status)
		return errors.Wrap(err, "unable to execute command")
	}

	l.Info("marking step as finished")
	if err := h.updateStepStatus(ctx, step.ID, startTS, models.AppInstallActionWorkflowRunStepStatusFinished); err != nil {
		return errors.Wrap(err, "unable to update status")
	}

	return nil
}
