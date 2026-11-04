package activities

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

// NOTE(jm): once install-deploys and sandbox-runs objects are updated to the composite status type, we will not need to
// use this type of update flow
type UpdateFlowStepTargetStatusRequest struct {
	StepID            string     `validate:"required"`
	Status            app.Status `validate:"required"`
	StatusDescription string
	Metadata          map[string]any
}

// @temporal-gen-v2 activity
func (a *Activities) PkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx context.Context, req UpdateFlowStepTargetStatusRequest) error {
	step, err := a.PkgWorkflowsFlowGetFlowsStep(ctx, GetFlowStepRequest{
		FlowStepID: req.StepID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to get step")
	}

	status := app.NewCompositeStatus(ctx, req.Status)
	if len(req.Metadata) > 0 {
		status.StatusHumanDescription = req.StatusDescription
	}

	switch step.StepTargetType {
	case "install_deploys":
		obj := &app.InstallDeploy{
			ID: step.StepTargetID,
		}
		return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			res := tx.Model(obj).Updates(app.InstallDeploy{
				Status:            app.InstallDeployStatus(req.Status),
				StatusDescription: req.StatusDescription,
				StatusV2:          status,
			})
			if res.Error != nil {
				return errors.Wrap(res.Error, "unable to update install_deploy")
			}
			if len(req.Metadata) > 0 {
				return generics.MergeJSONBMetadata(tx, obj, obj.ID, "status_v2", req.Metadata)
			}
			return nil
		})
	case "install_sandbox_runs":
		obj := &app.InstallSandboxRun{
			ID: step.StepTargetID,
		}
		return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			res := tx.Model(obj).Updates(app.InstallSandboxRun{
				Status:            app.SandboxRunStatus(req.Status),
				StatusDescription: req.StatusDescription,
				StatusV2:          status,
			})
			if res.Error != nil {
				return errors.Wrap(res.Error, "unable to update install_sandbox_run")
			}
			if len(req.Metadata) > 0 {
				return generics.MergeJSONBMetadata(tx, obj, obj.ID, "status_v2", req.Metadata)
			}
			return nil
		})
	case "install_stack_versions":
		obj := &app.InstallStackVersion{
			ID: step.StepTargetID,
		}
		res := a.db.WithContext(ctx).
			Model(obj).
			Updates(map[string]any{
				"status": status,
			})
		if res.Error != nil {
			return errors.Wrap(res.Error, "unable to update install_stack_version")
		}
	}

	return nil
}
