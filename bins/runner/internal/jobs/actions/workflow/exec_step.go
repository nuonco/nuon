package workflow

import (
	"context"

	"github.com/pkg/errors"
	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/git"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"

	"github.com/nuonco/nuon-runner-go/models"
)

func (h *handler) updateStepStatus(ctx context.Context, stepID string, status models.AppInstallActionWorkflowRunStepStatus) error {
	_, err := h.apiClient.UpdateInstallActionWorkflowRunStep(ctx, h.state.plan.InstallID, h.state.workflowCfg.ActionWorkflowID, stepID, &models.ServiceUpdateInstallActionWorkflowRunStepRequest{
		Status: models.AppInstallActionWorkflowRunStepStatusInDashProgress,
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

	if err := h.updateStepStatus(ctx, step.ID, models.AppInstallActionWorkflowRunStepStatusInDashProgress); err != nil {
		return err
	}

	// TODO(jm): fix this on the backend to use tokens, or whatever
	// should use the plan
	src := &git.Source{
		URL: "https://github.com/nuonco/actions",
		Ref: "jm/test",
	}
	if err := h.createExecEnv(ctx, l, src); err != nil {
		return errors.Wrap(err, "unable to create exec env")
	}

	return nil
}
