package workflow

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
)

func (h *handler) noopWorkflowSteps(ctx context.Context, steps []*models.AppInstallActionWorkflowRunStep, cfgs []*models.AppActionWorkflowStepConfig) error {
	for idx, step := range steps {
		if err := h.noopWorkflowStep(ctx, step, cfgs[idx]); err != nil {
			return errors.Wrap(err, "unable to exec noop step")
		}
	}
	return nil
}

func (h *handler) noopWorkflowStep(ctx context.Context, step *models.AppInstallActionWorkflowRunStep, cfg *models.AppActionWorkflowStepConfig) error {
	if err := h.updateStepStatus(ctx, step.ID, models.AppInstallActionWorkflowRunStepStatusError); err != nil {
		return errors.Wrap(err, "unable to mark step as NOOP")
	}

	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l = l.With(
		zap.String("workflow_step_name", cfg.Name),
		zap.String("step_run_id", step.ID),
	)
	l.Warn("step was not attempted because a previous workflow step already failed")

	return nil
}
