package activities

import (
	"context"
	"errors"
	"time"

	"go.temporal.io/sdk/temporal"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/driftresolution"
)

type UpdateFlowStepStartedAtRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field ID
func (a *Activities) PkgWorkflowsFlowUpdateFlowStepStartedAt(ctx context.Context, req UpdateFlowStepStartedAtRequest) error {
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var step app.WorkflowStep
		if err := tx.Where(app.WorkflowStep{ID: req.ID}).First(&step).Error; err != nil {
			return generics.TemporalGormError(err)
		}
		var wf app.Workflow
		if err := tx.Where(app.Workflow{ID: step.InstallWorkflowID}).First(&wf).Error; err != nil {
			return err
		}
		if target := driftresolution.Target(&wf); target != nil && step.QueueSignal != nil && step.QueueSignal.Signal != nil {
			switch step.QueueSignal.Signal.Type() {
			case "component-deploy-apply-plan", "component-teardown-apply-plan", "reprovision-sandbox-apply-plan", "generate-install-stack-version", "component-sync-image":
				if step.QueueSignal.Signal.Type() == "component-sync-image" && target.Kind != "image" {
					break
				}
				if err := driftresolution.AdmitApply(ctx, tx, &wf, step.ID); err != nil {
					var changed driftresolution.TargetChanged
					if errors.As(err, &changed) {
						return temporal.NewNonRetryableApplicationError(err.Error(), "resolution-target-changed", err)
					}
					return err
				}
			}
		}
		return tx.Model(&step).Updates(app.WorkflowStep{StartedAt: time.Now()}).Error
	})
}
