package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// NOTE(jm): once install-deploys and sandbox-runs objects are updated to the composite status type, we will not need to
// use this type of update flow
type UpdateFlowStepTargetStatusRequest struct {
	StepID            string     `validate:"required"`
	Status            app.Status `validate:"required"`
	StatusDescription string
}

// @temporal-gen activity
func (a *Activities) PkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx context.Context, req UpdateFlowStepTargetStatusRequest) error {
	step, err := a.PkgWorkflowsFlowGetFlowsStep(ctx, GetFlowStepRequest{
		FlowStepID: req.StepID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to get step")
	}

	switch step.StepTargetType {
	case "install_deploys":
		obj := &app.InstallDeploy{
			ID: step.StepTargetID,
		}
		res := a.db.WithContext(ctx).
			Model(obj).
			Updates(app.InstallDeploy{
				Status:            app.InstallDeployStatus(req.Status),
				StatusDescription: req.StatusDescription,
				StatusV2:          app.NewCompositeStatus(ctx, req.Status),
			})
		if res.Error != nil {
			return errors.Wrap(res.Error, "unable to update install_deploy")
		}
	case "install_sandbox_runs":
		obj := &app.InstallSandboxRun{
			ID: step.StepTargetID,
		}
		res := a.db.WithContext(ctx).
			Model(obj).
			Updates(app.InstallSandboxRun{
				Status:            app.SandboxRunStatus(req.Status),
				StatusDescription: req.StatusDescription,
				StatusV2:          app.NewCompositeStatus(ctx, req.Status),
			})
		if res.Error != nil {
			return errors.Wrap(res.Error, "unable to update install_sandbox_run")
		}
	}

	return nil
}
