package workflow

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/git"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"

	"github.com/nuonco/nuon-runner-go/models"
)

func (h *handler) updateStepStatus(ctx context.Context, stepID string, status models.AppInstallActionWorkflowRunStepStatus) error {
	_, err := h.apiClient.UpdateInstallActionWorkflowRunStep(ctx, h.state.plan.InstallID, h.state.workflowCfg.ActionWorkflowID, stepID, &models.ServiceUpdateInstallActionWorkflowRunStepRequest{
		Status: status,
	})
	if err != nil {
		return errors.Wrap(err, "unable to update step status")
	}

	return nil
}

func (h *handler) executeWorkflowStep(ctx context.Context, step *models.AppInstallActionWorkflowRunStep, cfg *models.AppActionWorkflowStepConfig, stepPlan *plantypes.ActionWorkflowRunStepPlan) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l = l.With(
		zap.String("workflow_step_name", cfg.Name),
		zap.String("step_run_id", step.ID),
	)

	if err := h.updateStepStatus(ctx, step.ID, models.AppInstallActionWorkflowRunStepStatusInDashProgress); err != nil {
		return errors.Wrap(err, "unable to update status")
	}

	// TODO(jm): fix this on the backend to use tokens, or whatever
	// should use the plan
	src := &git.Source{
		URL: "https://github.com/nuonco/actions",
		Ref: "main",
	}
	if err := h.createExecEnv(ctx, l, src); err != nil {
		h.updateStepStatus(ctx, step.ID, models.AppInstallActionWorkflowRunStepStatusError)
		return errors.Wrap(err, "unable to create exec env")
	}

	if err := h.execCommand(ctx, l, cfg, src); err != nil {
		h.updateStepStatus(ctx, step.ID, models.AppInstallActionWorkflowRunStepStatusError)
		return errors.Wrap(err, "unable to execute command")
	}

	l.Info("marking step as finished")
	if err := h.updateStepStatus(ctx, step.ID, models.AppInstallActionWorkflowRunStepStatusFinished); err != nil {
		return errors.Wrap(err, "unable to update status")
	}

	return nil
}
