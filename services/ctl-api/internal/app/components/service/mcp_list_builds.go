package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpListBuildsInput struct {
	ComponentID string `json:"component_id" jsonschema:"component ID to list builds for"`
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
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	var builds []app.ComponentBuild
	err = s.db.WithContext(ctx).
		Preload("ComponentConfigConnection").
		Preload("ComponentConfigConnection.Component").
		Where(app.ComponentBuild{OrgID: orgID}).
		Where("component_config_connection_id IN (?)",
			s.db.Model(&app.ComponentConfigConnection{}).
				Select("id").
				Where(app.ComponentConfigConnection{ComponentID: in.ComponentID}),
		).
		Order("created_at DESC").
		Limit(20).
		Find(&builds).Error
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list builds: %w", err)
	}

	out := make([]mcpBuildSummary, 0, len(builds))
	for _, b := range builds {
		summary := mcpBuildSummary{
			ID:        b.ID,
			Status:    string(b.Status),
			CreatedAt: b.CreatedAt.String(),
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

	return apiPkg.MCPJSONResult(out)
}
