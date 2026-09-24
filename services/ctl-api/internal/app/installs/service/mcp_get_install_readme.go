package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpGetInstallReadmeInput struct {
	Install string `json:"install" jsonschema:"install name or ID"`
}

type mcpGetInstallReadmeResult struct {
	InstallID   string   `json:"install_id"`
	InstallName string   `json:"install_name"`
	Readme      string   `json:"readme"`
	Warnings    []string `json:"warnings,omitempty"`
}

func (s *service) mcpGetInstallReadme(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetInstallReadmeInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.Install == "" {
		return nil, nil, fmt.Errorf("install is required")
	}

	install, err := s.findInstall(ctx, orgID, in.Install)
	if err != nil {
		return nil, nil, err
	}
	readme, err := s.renderInstallReadme(ctx, orgID, install.ID)
	if err != nil {
		return nil, nil, err
	}

	return apiPkg.MCPJSONResult(mcpGetInstallReadmeResult{
		InstallID:   install.ID,
		InstallName: install.Name,
		Readme:      readme.Rendered,
		Warnings:    readme.Warnings,
	})
}
