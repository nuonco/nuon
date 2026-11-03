package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpListComponentsInput struct {
	AppID  string `json:"app_id,omitempty" jsonschema:"filter components by app ID"`
	Limit  int    `json:"limit,omitempty" jsonschema:"maximum components to return (default 20, max 100)"`
	Offset int    `json:"offset,omitempty" jsonschema:"skip this many components; use next_offset from a previous response when has_more is true"`
}

type mcpComponentListItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	AppID     string `json:"app_id,omitempty"`
	Status    string `json:"status,omitempty"`
	CreatedAt string `json:"created_at"`
}

type mcpListComponentsResult struct {
	Components []mcpComponentListItem `json:"components"`
	HasMore    bool                   `json:"has_more"`
	Limit      int                    `json:"limit"`
	Offset     int                    `json:"offset"`
	NextOffset int                    `json:"next_offset,omitempty"`
}

func (s *service) mcpListComponents(ctx context.Context, _ *mcp.CallToolRequest, in mcpListComponentsInput) (*mcp.CallToolResult, any, error) {
	limit, offset, err := apiPkg.MCPListPage(in.Limit, in.Offset)
	if err != nil {
		return nil, nil, err
	}

	contract := require.ReadContract{Permission: permissions.AppRead, Collection: permissions.KindComponent}
	if in.AppID != "" {
		contract.Targets = []require.ReadTarget{{Name: "app_id", Kind: permissions.KindApp}}
	}
	components, err := require.ReadOperation(ctx, s.db, contract, map[string]string{"app_id": in.AppID}, func(ctx context.Context) ([]app.Component, error) {
		access, err := require.ReadAccessFromContext(ctx)
		if err != nil {
			return nil, err
		}
		tx, err := access.Query()
		if err != nil {
			return nil, err
		}
		if in.AppID != "" {
			target, err := access.Target("app_id")
			if err != nil {
				return nil, err
			}
			tx = tx.Where(&app.Component{AppID: target.ID})
		}
		var components []app.Component
		if err := tx.WithContext(ctx).Order("components.name ASC, components.id ASC").Limit(limit + 1).Offset(offset).Find(&components).Error; err != nil {
			return nil, fmt.Errorf("unable to list components: %w", err)
		}
		return components, nil
	})
	if err != nil {
		return nil, nil, err
	}

	components, hasMore := apiPkg.MCPClipList(components, limit)

	out := make([]mcpComponentListItem, 0, len(components))
	for _, c := range components {
		out = append(out, mcpComponentListItem{
			ID:        c.ID,
			Name:      c.Name,
			Type:      string(c.Type),
			AppID:     c.AppID,
			Status:    string(c.Status),
			CreatedAt: apiPkg.MCPTime(c.CreatedAt),
		})
	}

	return apiPkg.MCPJSONResult(mcpListComponentsResult{
		Components: out,
		HasMore:    hasMore,
		Limit:      limit,
		Offset:     offset,
		NextOffset: apiPkg.MCPNextOffset(offset, limit, hasMore),
	})
}
