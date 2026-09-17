package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpGetInstallHealthInput struct {
	Install string `json:"install" jsonschema:"install name or ID"`
}

type mcpInstallComponentHealth struct {
	InstallComponentID string `json:"install_component_id"`
	ComponentID        string `json:"component_id"`
	ComponentName      string `json:"component_name"`
	DeployStatus       string `json:"deploy_status,omitempty"`
	HealthStatus       string `json:"health_status,omitempty"`
	HealthDescription  string `json:"health_description,omitempty"`
}

type mcpGetInstallHealthResult struct {
	InstallID          string                      `json:"install_id"`
	InstallName        string                      `json:"install_name"`
	HealthStatus       string                      `json:"health_status,omitempty"`
	HealthDescription  string                      `json:"health_description,omitempty"`
	SandboxHealth      string                      `json:"sandbox_health,omitempty"`
	SandboxMessage     string                      `json:"sandbox_message,omitempty"`
	ClusterAccessError string                      `json:"cluster_access_error,omitempty"`
	LastReportedAt     string                      `json:"last_reported_at,omitempty"`
	Components         []mcpInstallComponentHealth `json:"components"`
}

func (s *service) mcpGetInstallHealth(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetInstallHealthInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.Install == "" {
		return nil, nil, fmt.Errorf("install is required")
	}
	if err := s.requireComponentHealthFeature(ctx, &app.Org{ID: orgID}); err != nil {
		return nil, nil, err
	}

	install, err := s.findInstall(ctx, orgID, in.Install)
	if err != nil {
		return nil, nil, err
	}

	components := append([]app.InstallComponent(nil), install.InstallComponents...)
	sort.Slice(components, func(i, j int) bool {
		return components[i].Component.Name < components[j].Component.Name
	})

	result := mcpGetInstallHealthResult{
		InstallID:          install.ID,
		InstallName:        install.Name,
		HealthStatus:       string(install.CompositeHealthStatus),
		HealthDescription:  install.CompositeHealthStatusDescription,
		SandboxHealth:      install.SandboxHealthStatus,
		SandboxMessage:     install.SandboxHealthMessage,
		ClusterAccessError: install.HealthClusterError,
		Components:         make([]mcpInstallComponentHealth, 0, len(components)),
	}
	if install.LastHealthReportAt != nil {
		result.LastReportedAt = apiPkg.MCPTimePtr(install.LastHealthReportAt)
	}
	for _, component := range components {
		result.Components = append(result.Components, mcpInstallComponentHealth{
			InstallComponentID: component.ID,
			ComponentID:        component.ComponentID,
			ComponentName:      component.Component.Name,
			DeployStatus:       string(component.Status),
			HealthStatus:       string(component.HealthStatus),
			HealthDescription:  component.HealthStatusDescription,
		})
	}

	return apiPkg.MCPJSONResult(result)
}
