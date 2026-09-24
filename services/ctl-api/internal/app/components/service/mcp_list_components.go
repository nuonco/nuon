package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
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
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	limit, offset, err := apiPkg.MCPListPage(in.Limit, in.Offset)
	if err != nil {
		return nil, nil, err
	}

	var components []app.Component
	tx := s.db.WithContext(ctx).
		Where(app.Component{OrgID: orgID}).
		Order("name ASC").
		Limit(limit + 1).
		Offset(offset)

	if in.AppID != "" {
		tx = tx.Where(app.Component{AppID: in.AppID})
	}

	if res := tx.Find(&components); res.Error != nil {
		return nil, nil, fmt.Errorf("unable to list components: %w", res.Error)
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
