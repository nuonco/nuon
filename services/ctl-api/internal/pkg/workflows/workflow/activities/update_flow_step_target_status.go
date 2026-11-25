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
	IncludeSuperseded bool
}

// @temporal-gen-v2 activity
func (a *Activities) PkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx context.Context, req UpdateFlowStepTargetStatusRequest) error {
	step, err := a.PkgWorkflowsFlowGetFlowsStep(ctx, GetFlowStepRequest{
		FlowStepID: req.StepID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to get step")
	}
	targetID := step.StepTargetID
	superseded := false
	if targetID == "" {
		if !req.IncludeSuperseded {
			return nil
		}
		targetID, err = a.supersededStepTargetID(ctx, step)
		if err != nil {
			return errors.Wrap(err, "unable to get superseded step target")
		}
		if targetID == "" {
			return nil
		}
		superseded = true
	}

	status := app.NewCompositeStatus(ctx, req.Status)
	if len(req.Metadata) > 0 {
		status.StatusHumanDescription = req.StatusDescription
	}

	switch step.StepTargetType {
	case "install_deploys":
		obj := &app.InstallDeploy{
			ID: targetID,
		}
		return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			q := tx.Model(obj)
			if superseded {
				q = q.Where("status = ?", app.InstallDeployStatusRetried)
			}
			res := q.Updates(app.InstallDeploy{
				Status:            app.InstallDeployStatus(req.Status),
				StatusDescription: req.StatusDescription,
				StatusV2:          status,
			})
			if res.Error != nil {
				return errors.Wrap(res.Error, "unable to update install_deploy")
			}
			if res.RowsAffected == 0 {
				return nil
			}
			if len(req.Metadata) > 0 {
				return generics.MergeJSONBMetadata(tx, obj, obj.ID, "status_v2", req.Metadata)
			}
			return nil
		})
	case "install_sandbox_runs":
		obj := &app.InstallSandboxRun{
			ID: targetID,
		}
		return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			q := tx.Model(obj)
			if superseded {
				q = q.Where("status = ?", app.SandboxRunStatusRetried)
			}
			res := q.Updates(app.InstallSandboxRun{
				Status:            app.SandboxRunStatus(req.Status),
				StatusDescription: req.StatusDescription,
				StatusV2:          status,
			})
			if res.Error != nil {
				return errors.Wrap(res.Error, "unable to update install_sandbox_run")
			}
			if res.RowsAffected == 0 {
				return nil
			}
			if len(req.Metadata) > 0 {
				return generics.MergeJSONBMetadata(tx, obj, obj.ID, "status_v2", req.Metadata)
			}
			return nil
		})
	case "install_stack_versions":
		if superseded {
			return nil
		}
		obj := &app.InstallStackVersion{
			ID: targetID,
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

func (a *Activities) supersededStepTargetID(ctx context.Context, step *app.WorkflowStep) (string, error) {
	if step.RetryIndex == 0 && step.GroupRetryIdx == 0 {
		return "", nil
	}
	var prev app.WorkflowStep
	res := a.db.WithContext(ctx).
		Where("install_workflow_id = ? AND group_idx = ? AND step_target_type = ?", step.InstallWorkflowID, step.GroupIdx, step.StepTargetType).
		Where("step_target_id IS NOT NULL AND step_target_id <> ''").
		Where("id <> ?", step.ID).
		Order("created_at DESC").
		Order("idx DESC").
		Limit(1).
		Find(&prev)
	if res.Error != nil {
		return "", res.Error
	}
	return prev.StepTargetID, nil
}
