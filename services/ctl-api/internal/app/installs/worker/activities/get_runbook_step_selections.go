package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetRunbookStepSelectionsRequest struct {
	InstallWorkflowID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallWorkflowID
func (a *Activities) GetRunbookStepSelections(ctx context.Context, req GetRunbookStepSelectionsRequest) ([]app.RunbookStepSelection, error) {
	var run app.InstallRunbookRun
	if err := a.db.WithContext(ctx).
		Select("step_selections").
		First(&run, "install_workflow_id = ?", req.InstallWorkflowID).Error; err != nil {
		return nil, fmt.Errorf("unable to get install runbook run: %w", err)
	}

	return run.StepSelections, nil
}
