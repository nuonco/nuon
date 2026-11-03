package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type mcpListBuildsInput struct {
	ComponentID string `json:"component_id" jsonschema:"component ID to list builds for"`
	Limit       int    `json:"limit,omitempty" jsonschema:"maximum builds to return (default 20, max 100)"`
	Offset      int    `json:"offset,omitempty" jsonschema:"skip this many builds; use next_offset from a previous response when has_more is true"`
}

type mcpBuildSummary struct {
	ID            string `json:"id"`
	ComponentID   string `json:"component_id,omitempty"`
	ComponentName string `json:"component_name,omitempty"`
	Status        string `json:"status"`
	GitRef        string `json:"git_ref,omitempty"`
	CreatedAt     string `json:"created_at"`
}

func (s *service) mcpListBuilds(ctx context.Context, _ *mcp.CallToolRequest, in mcpListBuildsInput) (*mcp.CallToolResult, any, error) {
	limit, offset, err := apiPkg.MCPListPage(in.Limit, in.Offset)
	if err != nil {
		return nil, nil, err
	}

	builds, err := require.ReadOperation(ctx, s.db, require.ReadContract{
		Permission: permissions.AppRead,
		Collection: permissions.KindComponentBuild,
		Targets:    []require.ReadTarget{{Name: "component_id", Kind: permissions.KindComponent}},
	}, map[string]string{"component_id": in.ComponentID}, func(ctx context.Context) ([]app.ComponentBuild, error) {
		access, err := require.ReadAccessFromContext(ctx)
		if err != nil {
			return nil, err
		}
		target, err := access.Target("component_id")
		if err != nil {
			return nil, err
		}
		query, err := access.Query()
		if err != nil {
			return nil, err
		}
		var builds []app.ComponentBuild
		err = query.WithContext(ctx).
			Preload("ComponentConfigConnection.Component").
			Where("component_builds.component_config_connection_id IN (?)", access.DB().Model(&app.ComponentConfigConnection{}).
				Scopes(scopes.WithDisableViews).Select("id").Where(&app.ComponentConfigConnection{ComponentID: target.ID, OrgID: access.OrgID()})).
			Order("component_builds.created_at DESC, component_builds.id DESC").
			Limit(limit + 1).Offset(offset).Find(&builds).Error
		return builds, err
	})
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list builds: %w", err)
	}
	builds, hasMore := apiPkg.MCPClipList(builds, limit)

	out := make([]mcpBuildSummary, 0, len(builds))
	for _, b := range builds {
		summary := mcpBuildSummary{
			ID:        b.ID,
			Status:    string(b.Status),
			CreatedAt: apiPkg.MCPTime(b.CreatedAt),
		}
		if b.GitRef != nil {
			summary.GitRef = *b.GitRef
		}
		if b.ComponentConfigConnection.Component.ID != "" {
			summary.ComponentID = b.ComponentConfigConnection.Component.ID
			summary.ComponentName = b.ComponentConfigConnection.Component.Name
		}
		out = append(out, summary)
	}

	return apiPkg.MCPJSONResult(struct {
		Builds     []mcpBuildSummary `json:"builds"`
		HasMore    bool              `json:"has_more"`
		Limit      int               `json:"limit"`
		Offset     int               `json:"offset"`
		NextOffset int               `json:"next_offset,omitempty"`
	}{out, hasMore, limit, offset, apiPkg.MCPNextOffset(offset, limit, hasMore)})
}
