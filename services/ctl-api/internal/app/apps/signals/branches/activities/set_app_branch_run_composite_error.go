package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/branchrunerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

type SetAppBranchRunCompositeErrorRequest struct {
	RunID  string `json:"run_id" validate:"required"`
	Detail string `json:"detail"`
}

// @temporal-gen-v2 activity
// @max-retries 3
func (a *Activities) SetAppBranchRunCompositeError(ctx context.Context, req SetAppBranchRunCompositeErrorRequest) error {
	var data *compositeerrors.CompositeErrorData
	if req.Detail != "" {
		var err error
		data, err = compositeerrors.New(
			&branchrunerrors.ConfigValidationFailedError{Detail: req.Detail},
			compositeerrors.WithSource("app_branch_runs", req.RunID),
		)
		if err != nil {
			return fmt.Errorf("unable to build app branch run composite error: %w", err)
		}
	}

	res := a.db.WithContext(ctx).
		Model(&app.AppBranchRun{ID: req.RunID}).
		Select("composite_error").
		Updates(app.AppBranchRun{CompositeError: data})
	if res.Error != nil {
		return fmt.Errorf("unable to set app branch run composite error: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no app branch run found for id %s: %w", req.RunID, gorm.ErrRecordNotFound)
	}
	return nil
}
