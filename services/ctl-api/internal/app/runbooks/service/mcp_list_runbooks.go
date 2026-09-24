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
	AppID  string `json:"app_id" jsonschema:"app ID to list runbooks for"`
	Limit  int    `json:"limit,omitempty" jsonschema:"maximum runbooks to return (default 20, max 100)"`
	Offset int    `json:"offset,omitempty" jsonschema:"skip this many runbooks; use next_offset from a previous response when has_more is true"`
}

type mcpRunbookListItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	CreatedAt   string `json:"created_at"`
}

type mcpListRunbooksResult struct {
	Runbooks   []mcpRunbookListItem `json:"runbooks"`
	HasMore    bool                 `json:"has_more"`
	Limit      int                  `json:"limit"`
	Offset     int                  `json:"offset"`
	NextOffset int                  `json:"next_offset,omitempty"`
}

func (s *service) mcpListRunbooks(ctx context.Context, _ *mcp.CallToolRequest, in mcpListRunbooksInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.AppID == "" {
		return nil, nil, fmt.Errorf("app_id is required")
	}

	limit, offset, err := apiPkg.MCPListPage(in.Limit, in.Offset)
	if err != nil {
		return nil, nil, err
	}

	var runbooks []app.Runbook
	err = s.db.WithContext(ctx).
		Where(app.Runbook{OrgID: orgID, AppID: in.AppID}).
		Order("created_at DESC").
		Limit(limit + 1).
		Offset(offset).
		Find(&runbooks).Error
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list runbooks: %w", err)
	}

	runbooks, hasMore := apiPkg.MCPClipList(runbooks, limit)

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

	return apiPkg.MCPJSONResult(mcpListRunbooksResult{
		Runbooks:   out,
		HasMore:    hasMore,
		Limit:      limit,
		Offset:     offset,
		NextOffset: apiPkg.MCPNextOffset(offset, limit, hasMore),
	})
}
