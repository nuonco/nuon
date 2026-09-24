package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpListAppsInput struct {
	Limit  int `json:"limit,omitempty" jsonschema:"maximum apps to return (default 20, max 100)"`
	Offset int `json:"offset,omitempty" jsonschema:"skip this many apps; use next_offset from a previous response when has_more is true"`
}

type mcpAppListItem struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status,omitempty"`
	CreatedAt   string   `json:"created_at"`
	Components  []string `json:"components,omitempty"`
}

type mcpListAppsResult struct {
	Apps       []mcpAppListItem `json:"apps"`
	HasMore    bool             `json:"has_more"`
	Limit      int              `json:"limit"`
	Offset     int              `json:"offset"`
	NextOffset int              `json:"next_offset,omitempty"`
}

func (s *service) mcpListApps(ctx context.Context, _ *mcp.CallToolRequest, in mcpListAppsInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	limit, offset, err := apiPkg.MCPListPage(in.Limit, in.Offset)
	if err != nil {
		return nil, nil, err
	}

	var apps []*app.App
	err = s.db.WithContext(ctx).
		Where(app.App{OrgID: orgID}).
		Preload("Components").
		Order("name ASC").
		Limit(limit + 1).
		Offset(offset).
		Find(&apps).Error
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list apps: %w", err)
	}

	apps, hasMore := apiPkg.MCPClipList(apps, limit)

	out := make([]mcpAppListItem, 0, len(apps))
	for _, a := range apps {
		a = mcpAppWithDerivedStatus(a)
		item := mcpAppListItem{
			ID:        a.ID,
			Name:      a.Name,
			Status:    string(a.Status),
			CreatedAt: apiPkg.MCPTime(a.CreatedAt),
		}
		if a.Description.Valid {
			item.Description = a.Description.String
		}
		for _, c := range a.Components {
			item.Components = append(item.Components, c.Name)
		}
		out = append(out, item)
	}

	return apiPkg.MCPJSONResult(mcpListAppsResult{
		Apps:       out,
		HasMore:    hasMore,
		Limit:      limit,
		Offset:     offset,
		NextOffset: apiPkg.MCPNextOffset(offset, limit, hasMore),
	})
}
