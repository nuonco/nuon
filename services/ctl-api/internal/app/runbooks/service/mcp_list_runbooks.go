package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type mcpListRunbooksInput struct {
	AppID string `json:"app_id" jsonschema:"app ID to list runbooks for"`
}

func (s *service) mcpListRunbooks(ctx context.Context, _ *mcp.CallToolRequest, in mcpListRunbooksInput) (*mcp.CallToolResult, any, error) {
	orgID := keys.OrgIDFromContext(ctx)

	var runbooks []app.Runbook
	err := s.db.WithContext(ctx).
		Where(app.Runbook{OrgID: orgID, AppID: in.AppID}).
		Order("created_at DESC").
		Find(&runbooks).Error
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list runbooks: %w", err)
	}

	return apiPkg.MCPJSONResult(runbooks)
}
