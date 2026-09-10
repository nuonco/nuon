package activities

import (
	"context"
	stderrors "errors"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetFlowMetaRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field ID
// @local
// @local-retry-policy-max-attempts 3
//
// PkgWorkflowsFlowGetFlowMeta returns a workflow without its Steps. Callers that
// only need the flow's identity and routing fields use this instead of
// PkgWorkflowsFlowGetFlow, whose Steps preload makes its cost grow with step
// count — a workflow with hundreds of steps pays for all of them to read OwnerID.
func (a *Activities) PkgWorkflowsFlowGetFlowMeta(ctx context.Context, req GetFlowMetaRequest) (*app.Workflow, error) {
	wf := app.Workflow{
		ID: req.ID,
	}
	if res := a.db.WithContext(ctx).
		Preload("Org", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name")
		}).
		First(&wf, "id = ?", req.ID); res.Error != nil {
		if stderrors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, temporal.NewNonRetryableApplicationError("workflow not found", "not found", res.Error)
		}
		return nil, errors.Wrap(res.Error, "unable to get install workflow")
	}

	a.resolveFlowOwnerName(ctx, &wf)

	return &wf, nil
}

// resolveFlowOwnerName fills OwnerName from the matching polymorphic owner table
// with one PK lookup. Best-effort: errors leave OwnerName empty.
func (a *Activities) resolveFlowOwnerName(ctx context.Context, wf *app.Workflow) {
	if wf.OwnerID == "" {
		return
	}

	var ownerTable string
	switch wf.OwnerType {
	case "installs":
		ownerTable = "installs"
	case "apps":
		ownerTable = "apps"
	case "app_branches":
		ownerTable = "app_branches"
	}
	if ownerTable == "" {
		return
	}

	_ = a.db.WithContext(ctx).
		Table(ownerTable).
		Select("name").
		Where("id = ?", wf.OwnerID).
		Scan(&wf.OwnerName).Error
}
