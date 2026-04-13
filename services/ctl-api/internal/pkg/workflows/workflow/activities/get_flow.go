package activities

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetFlowRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field ID
func (a *Activities) PkgWorkflowsFlowGetFlow(ctx context.Context, req GetFlowRequest) (*app.Workflow, error) {
	wf := app.Workflow{
		ID: req.ID,
	}
	if res := a.db.WithContext(ctx).
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("group_idx, group_retry_idx, idx, created_at asc")
		}).
		First(&wf, "id = ?", req.ID); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install workflow")
	}

	return &wf, nil
}
