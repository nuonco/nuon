package activities

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/actions/actionerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

type SetActionWorkflowRunPreparationCompositeErrorRequest struct {
	RunID  string `validate:"required"`
	Detail string
}

// @temporal-gen-v2 activity
// @max-retries 3
func (a *Activities) SetActionWorkflowRunPreparationCompositeError(ctx context.Context, req SetActionWorkflowRunPreparationCompositeErrorRequest) error {
	l := temporalzap.GetActivityLogger(ctx).With(zap.String("action_workflow_run_id", req.RunID))

	var data *compositeerrors.CompositeErrorData
	if req.Detail != "" {
		var err error
		data, err = compositeerrors.New(
			&actionerrors.PreparationFailedError{Detail: req.Detail},
			compositeerrors.WithSource("install_action_workflow_runs", req.RunID),
		)
		if err != nil {
			return fmt.Errorf("unable to build action workflow run preparation composite error: %w", err)
		}
	}

	res := a.db.WithContext(ctx).
		Model(&app.InstallActionWorkflowRun{ID: req.RunID}).
		Select("composite_error").
		Updates(app.InstallActionWorkflowRun{CompositeError: data})
	if res.Error != nil {
		return fmt.Errorf("unable to set action workflow run preparation composite error: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no action workflow run found for id %s: %w", req.RunID, gorm.ErrRecordNotFound)
	}

	l.Info("updated action workflow run preparation composite error")
	return nil
}
