package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpListAppsInput struct {
	Limit  int    `json:"limit,omitempty" jsonschema:"maximum apps to return (default 20, max 100)"`
	Offset int    `json:"offset,omitempty" jsonschema:"skip this many apps; use next_offset from a previous response when has_more is true"`
	Query  string `json:"query,omitempty" jsonschema:"literal case-insensitive app name or ID search"`
}

type mcpListAppsResult struct {
	Apps       []AppListItem `json:"apps"`
	HasMore    bool          `json:"has_more"`
	Limit      int           `json:"limit"`
	Offset     int           `json:"offset"`
	NextOffset int           `json:"next_offset,omitempty"`
}

func (s *service) mcpListApps(ctx context.Context, _ *mcp.CallToolRequest, in mcpListAppsInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Org(ctx)
	if err != nil {
		return nil, nil, err
	}

	limit, offset, err := apiPkg.MCPListPage(in.Limit, in.Offset)
	if err != nil {
		return nil, nil, err
	}

	apps, err := s.getCustomerApps(ctx, orgID, in.Query, limit, offset)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list apps: %w", err)
	}

	apps, hasMore := apiPkg.MCPClipList(apps, limit)

	return apiPkg.MCPJSONResult(mcpListAppsResult{
		Apps:       apps,
		HasMore:    hasMore,
		Limit:      limit,
		Offset:     offset,
		NextOffset: apiPkg.MCPNextOffset(offset, limit, hasMore),
	})
}
