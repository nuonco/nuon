package activities

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/lifecyclephase"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type UpdateFlowFinishedAtRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field ID
func (a *Activities) PkgWorkflowsFlowUpdateFlowFinishedAt(ctx context.Context, req UpdateFlowFinishedAtRequest) error {
	var runner app.Workflow
	if err := a.db.WithContext(ctx).Where("id = ?", req.ID).Take(&runner).Error; err != nil {
		return generics.TemporalGormError(gorm.ErrRecordNotFound)
	}
	runner.FinishedAt = time.Now()
	if err := a.db.WithContext(ctx).Save(&runner).Error; err != nil {
		return generics.TemporalGormError(gorm.ErrRecordNotFound)
	}

	if runner.OwnerType == "installs" {
		if newPhase := a.installLifecycleTransition(&runner); newPhase != nil {
			var install app.Install
			if err := a.db.WithContext(ctx).Where(app.Install{ID: runner.OwnerID}).First(&install).Error; err == nil {
				transitioned := lifecyclephase.Transition(install.LifecyclePhase, newPhase.Phase, newPhase.Description)
				a.db.WithContext(ctx).Model(&app.Install{ID: runner.OwnerID}).Updates(map[string]any{
					"lifecycle_phase": transitioned,
				})
			}
		}
	}

	return nil
}

func (a *Activities) installLifecycleTransition(wf *app.Workflow) *lifecyclephase.LifecyclePhase {
	failed := wf.Status.Status == app.StatusError || wf.Status.Status == app.StatusCancelled

	switch wf.Type {
	case app.WorkflowTypeProvision:
		phase := lifecyclephase.Provisioned
		description := "Provision workflow completed"
		if failed {
			description = "Provision workflow failed"
		}
		return &lifecyclephase.LifecyclePhase{Phase: phase, Description: description}
	case app.WorkflowTypeDeprovision:
		description := "Deprovision workflow completed"
		if failed {
			description = "Deprovision workflow failed"
		}
		return &lifecyclephase.LifecyclePhase{Phase: lifecyclephase.Deprovisioned, Description: description}
	}
	return nil
}
