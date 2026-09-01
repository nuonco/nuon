package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpGetRunbookInput struct {
	RunbookID string `json:"runbook_id" jsonschema:"runbook ID"`
}

func (s *service) mcpGetRunbook(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetRunbookInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	var runbook app.Runbook
	err = s.db.WithContext(ctx).
		Preload("RunbookConfig").
		Where(app.Runbook{OrgID: orgID}).
		Where("id = ?", in.RunbookID).
		First(&runbook).Error
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find runbook %q: %w", in.RunbookID, err)
	}

	return apiPkg.MCPJSONResult(runbook)
}
