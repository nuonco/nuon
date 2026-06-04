package activities

import (
	"context"
	"time"

	"gorm.io/gorm"

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
		if newLifecycle := a.installLifecycleTransition(&runner); newLifecycle != nil {
			var install app.Install
			if err := a.db.WithContext(ctx).Where(app.Install{ID: runner.OwnerID}).First(&install).Error; err == nil {
				existing := install.LifecycleStatus
				existing.History = nil
				newLifecycle.CreatedAtTS = time.Now().Unix()
				newLifecycle.History = append([]app.CompositeStatus{existing}, install.LifecycleStatus.History...)
				if len(newLifecycle.History) > 25 {
					newLifecycle.History = newLifecycle.History[:25]
				}
				a.db.WithContext(ctx).Model(&app.Install{ID: runner.OwnerID}).Updates(map[string]any{
					"lifecycle_status": newLifecycle,
				})
			}
		}
	}

	return nil
}

func (a *Activities) installLifecycleTransition(wf *app.Workflow) *app.CompositeStatus {
	failed := wf.Status.Status == app.StatusError || wf.Status.Status == app.StatusCancelled

	switch wf.Type {
	case app.WorkflowTypeProvision:
		status := app.InstallLifecycleStatusProvisioned
		description := "Install has been provisioned"
		if failed {
			description = "Install provision failed"
		}
		return &app.CompositeStatus{
			Status:                 app.Status(status),
			StatusHumanDescription: description,
		}
	case app.WorkflowTypeDeprovision:
		description := "Install has been deprovisioned"
		if failed {
			description = "Install deprovision failed"
		}
		return &app.CompositeStatus{
			Status:                 app.Status(app.InstallLifecycleStatusDeprovisioned),
			StatusHumanDescription: description,
		}
	}
	return nil
}
