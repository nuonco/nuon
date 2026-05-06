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
		// Preload Org with a column-restricted SELECT so the lifecycle
		// hook can stamp org_name onto webhook payloads without an extra
		// query at emit time. We deliberately fetch only id + name to
		// avoid pulling the rest of the (wide) orgs row.
		Preload("Org", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name")
		}).
		First(&wf, "id = ?", req.ID); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install workflow")
	}

	// Resolve the polymorphic owner's display name with one cheap PK lookup.
	// This runs once per Validate(), not per event, and lets the lifecycle
	// hook stamp owner_name onto webhook payloads without a per-event query.
	// Best-effort: errors leave OwnerName empty.
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
