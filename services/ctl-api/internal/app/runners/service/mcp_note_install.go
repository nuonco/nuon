package service

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *service) mcpWorkflowInstallID(ctx context.Context, orgID, workflowID string) string {
	if workflowID == "" {
		return ""
	}
	var workflow app.Workflow
	if err := s.db.WithContext(ctx).
		Select("owner_id", "owner_type").
		Where("id = ? AND org_id = ?", workflowID, orgID).
		First(&workflow).Error; err != nil {
		return ""
	}
	if workflow.OwnerType != "installs" {
		return ""
	}
	return workflow.OwnerID
}

func (s *service) noteMCPInstallForDeploy(ctx context.Context, orgID string, deploy app.InstallDeploy) {
	if deploy.InstallComponentID == "" {
		return
	}
	var component app.InstallComponent
	if err := s.db.WithContext(ctx).
		Select("install_id").
		Where("id = ? AND org_id = ?", deploy.InstallComponentID, orgID).
		First(&component).Error; err != nil {
		return
	}
	ctx = cctx.SetInstallIDContext(ctx, component.InstallID)
}
