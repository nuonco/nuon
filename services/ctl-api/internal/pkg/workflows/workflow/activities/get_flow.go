package activities

import (
	"context"
	stderrors "errors"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
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
		Preload("Org", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name")
		}).
		First(&wf, "id = ?", req.ID); res.Error != nil {
		if stderrors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, temporal.NewNonRetryableApplicationError("workflow not found", "not found", res.Error)
		}
		return nil, errors.Wrap(res.Error, "unable to get install workflow")
	}

	if wf.OwnerID != "" {
		var ownerTable string
		switch wf.OwnerType {
		case "installs":
			ownerTable = "installs"
		case "apps":
			ownerTable = "apps"
		case "app_branches":
			ownerTable = "app_branches"
		}
		if ownerTable != "" {
			_ = a.db.WithContext(ctx).
				Table(ownerTable).
				Select("name").
				Where("id = ?", wf.OwnerID).
				Scan(&wf.OwnerName).Error
		}
	}

	return &wf, nil
}
