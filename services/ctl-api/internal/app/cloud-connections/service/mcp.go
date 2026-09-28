package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpListInput struct {
	Limit  int `json:"limit,omitempty" jsonschema:"maximum connections to return (default 20, max 100)"`
	Offset int `json:"offset,omitempty" jsonschema:"skip this many connections"`
}

type mcpListResult struct {
	Connections []ConnectionResponse `json:"connections"`
	HasMore     bool                 `json:"has_more"`
	Limit       int                  `json:"limit"`
	Offset      int                  `json:"offset"`
	NextOffset  int                  `json:"next_offset,omitempty"`
}

type mcpConnectionInput struct {
	ConnectionID string `json:"connection_id" jsonschema:"cloud connection ID"`
}

type mcpCreateInput struct {
	Name          string                    `json:"name" jsonschema:"connection name"`
	Platform      app.CloudPlatform         `json:"platform,omitempty" jsonschema:"AWS only (defaults to aws),enum=aws"`
	TargetID      string                    `json:"target_id" jsonschema:"AWS account ID"`
	Principal     string                    `json:"principal" jsonschema:"AWS IAM role ARN"`
	DefaultRegion string                    `json:"default_region,omitempty" jsonschema:"AWS region, defaults to us-east-1"`
	Preset        app.CloudConnectionPreset `json:"preset" jsonschema:"Access preset. For custom attach your own permissions policy,enum=stacks,enum=custom"`
}

type mcpVerifyInput struct {
	ConnectionID string `json:"connection_id" jsonschema:"cloud connection ID"`
}

func (s *service) mcpList(ctx context.Context, _ *mcp.CallToolRequest, in mcpListInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}
	limit, offset, err := apiPkg.MCPListPage(in.Limit, in.Offset)
	if err != nil {
		return nil, nil, err
	}
	var connections []app.CloudConnection
	if err := s.db.WithContext(ctx).Where(app.CloudConnection{OrgID: orgID}).Order("created_at DESC").Limit(limit + 1).Offset(offset).Find(&connections).Error; err != nil {
		return nil, nil, fmt.Errorf("list cloud connections: %w", err)
	}
	connections, hasMore := apiPkg.MCPClipList(connections, limit)
	responses := make([]ConnectionResponse, 0, len(connections))
	for i := range connections {
		response, err := s.response(ctx, &connections[i])
		if err != nil {
			return nil, nil, err
		}
		responses = append(responses, response)
	}
	return apiPkg.MCPJSONResult(mcpListResult{Connections: responses, HasMore: hasMore, Limit: limit, Offset: offset, NextOffset: apiPkg.MCPNextOffset(offset, limit, hasMore)})
}

func (s *service) mcpGet(ctx context.Context, _ *mcp.CallToolRequest, in mcpConnectionInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}
	connection, err := s.getContext(ctx, orgID, in.ConnectionID)
	if err != nil {
		return nil, nil, err
	}
	response, err := s.response(ctx, connection)
	if err != nil {
		return nil, nil, err
	}
	return apiPkg.MCPJSONResult(response)
}

func (s *service) mcpCreate(ctx context.Context, _ *mcp.CallToolRequest, in mcpCreateInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Write(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.Platform == "" {
		in.Platform = app.CloudPlatformAWS
	}
	connection := app.CloudConnection{OrgID: orgID, Name: in.Name, Platform: in.Platform, TargetID: in.TargetID, Principal: in.Principal, DefaultRegion: in.DefaultRegion, Preset: in.Preset}
	if err := validateConnection(&connection); err != nil {
		return nil, nil, err
	}
	if err := s.db.WithContext(ctx).Create(&connection).Error; err != nil {
		return nil, nil, fmt.Errorf("create cloud connection: %w", err)
	}
	if _, err := s.helpers.EnsureConnectionQueue(ctx, &connection); err != nil {
		return nil, nil, err
	}
	response, err := s.response(ctx, &connection)
	if err != nil {
		return nil, nil, err
	}
	return apiPkg.MCPJSONResult(response)
}

func (s *service) mcpVerify(ctx context.Context, _ *mcp.CallToolRequest, in mcpVerifyInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Write(ctx)
	if err != nil {
		return nil, nil, err
	}
	connection, err := s.verify(ctx, orgID, in.ConnectionID)
	if err != nil {
		return nil, nil, err
	}
	response, err := s.response(ctx, connection)
	if err != nil {
		return nil, nil, err
	}
	return apiPkg.MCPJSONResult(response)
}

func (s *service) mcpDelete(ctx context.Context, _ *mcp.CallToolRequest, in mcpConnectionInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Write(ctx)
	if err != nil {
		return nil, nil, err
	}
	if err := s.delete(ctx, orgID, in.ConnectionID); err != nil {
		return nil, nil, err
	}
	return apiPkg.MCPJSONResult(map[string]any{"deleted": true, "connection_id": in.ConnectionID})
}
