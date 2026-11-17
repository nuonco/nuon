package service

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

func (s *service) mcpInstallPoint(ctx context.Context, id string, operation permissions.PermissionID) error {
	return require.MCPPoint(ctx, s.db, access.Installs, operation, id)
}

func (s *service) mcpWorkflowPoint(ctx context.Context, id string, operation permissions.PermissionID) error {
	return require.MCPPoint(ctx, s.db, access.Workflows, operation, id)
}

func (s *service) mcpApprovalContinuation(ctx context.Context, orgID string, approval app.WorkflowStepApproval) error {
	if !authz.HasAuthority(ctx) {
		return nil
	}
	step := approval.InstallWorkflowStep
	if step.ID == "" || step.ID != approval.InstallWorkflowStepID || step.OrgID != orgID {
		return gorm.ErrRecordNotFound
	}
	return s.mcpWorkflowContinuation(ctx, orgID, step.InstallWorkflowID, permissions.WorkflowApprove)
}

func (s *service) mcpWorkflowCollection(ctx context.Context) func(*gorm.DB) *gorm.DB {
	return require.MCPCollection(ctx, s.db, access.Workflows, permissions.WorkflowInspect)
}

func (s *service) mcpDeployCollection(ctx context.Context, orgID string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if !authz.HasAuthority(ctx) {
			return db
		}
		installs := s.db.WithContext(ctx).Model(&app.Install{}).Select("id").
			Scopes(require.MCPCollection(ctx, s.db, access.Installs, permissions.InstallHistoryRead))
		ids := s.db.Session(&gorm.Session{NewDB: true}).Table("(?) AS d", access.Deploys.Query(s.db, orgID)).Select("d.id").Where("d.install_id IN (?)", installs)
		return db.Where("install_deploys.id IN (?)", ids)
	}
}

func (s *service) mcpApprovalCollection(ctx context.Context, orgID string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if !authz.HasAuthority(ctx) {
			return db
		}
		workflows := s.db.WithContext(ctx).Model(&app.Workflow{}).Select("install_workflows.id").Scopes(s.mcpWorkflowCollection(ctx))
		steps := s.db.WithContext(ctx).Model(&app.WorkflowStep{}).Select("id").Where(app.WorkflowStep{OrgID: orgID}).Where("install_workflow_id IN (?)", workflows)
		return db.Where("install_workflow_step_id IN (?)", steps)
	}
}
