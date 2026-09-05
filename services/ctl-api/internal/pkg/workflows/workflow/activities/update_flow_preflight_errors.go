package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

type UpdateFlowPreflightErrorsRequest struct {
	FlowID          string `validate:"required"`
	PreflightErrors []*compositeerrors.CompositeErrorData
}

// @temporal-gen-v2 activity
func (a *Activities) PkgWorkflowsFlowUpdateFlowPreflightErrors(ctx context.Context, req UpdateFlowPreflightErrorsRequest) error {
	if req.PreflightErrors == nil {
		req.PreflightErrors = []*compositeerrors.CompositeErrorData{}
	}

	res := a.db.WithContext(ctx).
		Model(&app.Workflow{ID: req.FlowID}).
		Select("preflight_errors").
		Updates(app.Workflow{PreflightErrors: req.PreflightErrors})
	if res.Error != nil {
		return fmt.Errorf("unable to update workflow preflight errors: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no workflow found for id %s: %w", req.FlowID, gorm.ErrRecordNotFound)
	}
	return nil
}
