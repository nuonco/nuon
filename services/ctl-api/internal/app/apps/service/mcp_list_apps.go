package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpListAppsInput struct{}

type mcpAppListItem struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status,omitempty"`
	CreatedAt   string   `json:"created_at"`
	Components  []string `json:"components,omitempty"`
}

func (s *service) mcpListApps(ctx context.Context, _ *mcp.CallToolRequest, _ mcpListAppsInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	var apps []*app.App
	org := &app.Org{ID: orgID}

	err = s.db.WithContext(ctx).
		Preload("Components").
		Order("apps.name ASC").
		Model(org).Association("Apps").Find(&apps)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list apps: %w", err)
	}

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

	return apiPkg.MCPJSONResult(out)
}
