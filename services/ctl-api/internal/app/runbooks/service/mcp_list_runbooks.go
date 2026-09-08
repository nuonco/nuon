package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpListRunbooksInput struct {
	AppID string `json:"app_id" jsonschema:"app ID to list runbooks for"`
}

type mcpRunbookListItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	CreatedAt   string `json:"created_at"`
}

func (s *service) mcpListRunbooks(ctx context.Context, _ *mcp.CallToolRequest, in mcpListRunbooksInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	var runbooks []app.Runbook
	err = s.db.WithContext(ctx).
		Where(app.Runbook{OrgID: orgID, AppID: in.AppID}).
		Order("created_at DESC").
		Find(&runbooks).Error
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list runbooks: %w", err)
	}

	out := make([]mcpRunbookListItem, 0, len(runbooks))
	for _, r := range runbooks {
		out = append(out, mcpRunbookListItem{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Status:      string(r.Status),
			CreatedAt:   apiPkg.MCPTime(r.CreatedAt),
		})
	}

	return apiPkg.MCPJSONResult(out)
}
