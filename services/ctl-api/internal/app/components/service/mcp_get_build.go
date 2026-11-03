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

type mcpGetBuildInput struct {
	BuildID string `json:"build_id" jsonschema:"build ID"`
}

type mcpBuildDetail struct {
	ID                string `json:"id"`
	ComponentID       string `json:"component_id,omitempty"`
	ComponentName     string `json:"component_name,omitempty"`
	Status            string `json:"status"`
	StatusDescription string `json:"status_description,omitempty"`
	GitRef            string `json:"git_ref,omitempty"`
	SourceRef         string `json:"source_ref,omitempty"`
	LogStreamID       string `json:"log_stream_id,omitempty"`
	CreatedAt         string `json:"created_at"`
}

func (s *service) mcpGetBuild(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetBuildInput) (*mcp.CallToolResult, any, error) {
	build, err := require.ReadOperation(ctx, s.db, require.ReadContract{
		Permission: permissions.AppRead,
		Collection: permissions.KindComponentBuild,
		Targets:    []require.ReadTarget{{Name: "build_id", Kind: permissions.KindComponentBuild}},
	}, map[string]string{"build_id": in.BuildID}, func(ctx context.Context) (*app.ComponentBuild, error) {
		access, err := require.ReadAccessFromContext(ctx)
		if err != nil {
			return nil, err
		}
		target, err := access.Target("build_id")
		if err != nil {
			return nil, err
		}
		return s.getComponentBuild(ctx, "", target.ID)
	})
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find build %q: %w", in.BuildID, err)
	}

	detail := mcpBuildDetail{
		ID:                build.ID,
		Status:            string(build.Status),
		StatusDescription: build.StatusDescription,
		SourceRef:         build.SourceRef,
		CreatedAt:         apiPkg.MCPTime(build.CreatedAt),
	}
	if build.GitRef != nil {
		detail.GitRef = *build.GitRef
	}
	if build.ComponentConfigConnection.Component.ID != "" {
		detail.ComponentID = build.ComponentConfigConnection.Component.ID
		detail.ComponentName = build.ComponentConfigConnection.Component.Name
	}
	if build.LogStream.ID != "" {
		detail.LogStreamID = build.LogStream.ID
	}

	return apiPkg.MCPJSONResult(detail)
}
