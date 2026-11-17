package service

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

func (s *service) mcpAppPoint(ctx context.Context, id string, operation permissions.PermissionID) error {
	return require.MCPPoint(ctx, s.db, access.Apps, operation, id)
}

func (s *service) mcpAppCollection(ctx context.Context) func(*gorm.DB) *gorm.DB {
	return require.MCPCollection(ctx, s.db, access.Apps, permissions.AppInspect)
}

func (s *service) mcpBranchPoint(ctx context.Context, id string) error {
	return require.MCPPoint(ctx, s.db, access.Branches, permissions.AppBranchRead, id)
}

func (s *service) mcpBranchCollection(ctx context.Context) func(*gorm.DB) *gorm.DB {
	return require.MCPCollection(ctx, s.db, access.Branches, permissions.AppBranchRead)
}

func (s *service) mcpBranchRunCollection(ctx context.Context) func(*gorm.DB) *gorm.DB {
	return require.MCPCollection(ctx, s.db, access.BranchRuns, permissions.AppBranchRead)
}
