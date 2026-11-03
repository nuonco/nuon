package activities

import (
	"context"
	"time"

	"go.temporal.io/sdk/temporal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type UpdateFlowStepStartedAtRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field ID
func (a *Activities) PkgWorkflowsFlowUpdateFlowStepStartedAt(ctx context.Context, req UpdateFlowStepStartedAtRequest) error {
	if req.ID == "" {
		return temporal.NewNonRetryableApplicationError("step ID is required", authz.WorkflowAuthorizationDenied, nil)
	}
	var initial app.WorkflowStep
	if err := a.db.WithContext(ctx).Where(app.WorkflowStep{ID: req.ID}).Take(&initial).Error; err != nil {
		return generics.TemporalGormError(err)
	}
	decision, err := authz.New(authz.Params{DB: a.db}).AdmitWorkflow(ctx, initial.InstallWorkflowID, func(tx *gorm.DB, wf *app.Workflow) error {
		var step app.WorkflowStep
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.WorkflowStep{
			ID: req.ID, InstallWorkflowID: wf.ID, OrgID: wf.OrgID, OwnerID: wf.OwnerID, OwnerType: wf.OwnerType,
		}).Take(&step).Error; err != nil {
			return err
		}
		if !step.FinishedAt.IsZero() || (step.Status.Status != app.StatusPending && step.Status.Status != app.StatusQueued &&
			step.Status.Status != app.StatusNotAttempted && step.Status.Status != app.StatusInProgress) {
			return temporal.NewNonRetryableApplicationError("step is not executable", authz.WorkflowAuthorizationDenied, nil)
		}
		if !step.StartedAt.IsZero() {
			return nil
		}
		return tx.Model(&step).Update("started_at", time.Now()).Error
	})
	if err != nil {
		return generics.TemporalGormError(err)
	}
	if !decision.Allowed {
		return temporal.NewNonRetryableApplicationError("workflow authorization failed; submit a new request", authz.WorkflowAuthorizationDenied, nil, decision.Reason)
	}
	return nil
}
