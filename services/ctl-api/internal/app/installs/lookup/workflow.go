package lookup

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func Workflow(ctx context.Context, db *gorm.DB, orgID, workflowID string) (*app.Workflow, error) {
	if workflow, ok := cctx.ResolvedWorkflowFromContext(ctx, orgID, workflowID); ok {
		return workflow, nil
	}

	var workflow app.Workflow
	if err := db.WithContext(ctx).
		Where(&app.Workflow{ID: workflowID, OrgID: orgID}, "ID", "OrgID").
		First(&workflow).Error; err != nil {
		return nil, err
	}
	return &workflow, nil
}
