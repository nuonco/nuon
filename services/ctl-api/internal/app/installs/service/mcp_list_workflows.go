package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpListWorkflowsInput struct {
	InstallID string `json:"install_id" jsonschema:"install ID to list workflows for"`
}

func (s *service) mcpListWorkflows(ctx context.Context, _ *mcp.CallToolRequest, in mcpListWorkflowsInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	var workflows []app.Workflow
	err = s.db.WithContext(ctx).
		Where(app.Workflow{OrgID: orgID, OwnerID: in.InstallID}).
		Order("created_at DESC").
		Limit(20).
		Find(&workflows).Error
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list workflows: %w", err)
	}

	return apiPkg.MCPJSONResult(workflows)
}
