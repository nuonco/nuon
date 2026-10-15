package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type mcpListInstallResourcesInput struct {
	Install            string `json:"install" jsonschema:"install name or ID"`
	InstallComponentID string `json:"install_component_id,omitempty" jsonschema:"filter by install component ID"`
	Kind               string `json:"kind,omitempty" jsonschema:"filter by resource kind, such as Deployment"`
	Namespace          string `json:"namespace,omitempty" jsonschema:"filter by namespace"`
	Health             string `json:"health,omitempty" jsonschema:"filter by health: healthy, progressing, degraded, unhealthy, or unknown"`
	Provider           string `json:"provider,omitempty" jsonschema:"filter by provider: kubernetes, aws, gcp, azure, probe, or custom"`
	Limit              int    `json:"limit,omitempty" jsonschema:"maximum resources to return (default 20, max 100)"`
	Offset             int    `json:"offset,omitempty" jsonschema:"skip this many resources; use next_offset from a previous response when has_more is true"`
}

type mcpInstallResource struct {
	InstallComponentID string `json:"install_component_id,omitempty"`
	ComponentID        string `json:"component_id,omitempty"`
	ComponentName      string `json:"component_name,omitempty"`
	Source             string `json:"source"`
	OwnerName          string `json:"owner_name,omitempty"`
	Provider           string `json:"provider"`
	APIGroup           string `json:"api_group,omitempty"`
	Kind               string `json:"kind"`
	Namespace          string `json:"namespace,omitempty"`
	Name               string `json:"name"`
	Health             string `json:"health,omitempty"`
	Message            string `json:"message,omitempty"`
	NativeStatus       string `json:"native_status,omitempty"`
	ObservedAt         string `json:"observed_at"`
	RemovedFromConfig  bool   `json:"removed_from_config,omitempty"`
}

type mcpListInstallResourcesResult struct {
	InstallID   string               `json:"install_id"`
	InstallName string               `json:"install_name"`
	Resources   []mcpInstallResource `json:"resources"`
	HasMore     bool                 `json:"has_more"`
	Limit       int                  `json:"limit"`
	Offset      int                  `json:"offset"`
	NextOffset  int                  `json:"next_offset,omitempty"`
	NextActions []mcpNextAction      `json:"next_actions"`
}

func (s *service) mcpListInstallResources(ctx context.Context, _ *mcp.CallToolRequest, in mcpListInstallResourcesInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.Install == "" {
		return nil, nil, fmt.Errorf("install is required")
	}
	limit, offset, err := apiPkg.MCPListPage(in.Limit, in.Offset)
	if err != nil {
		return nil, nil, err
	}

	install, err := s.findInstall(ctx, orgID, in.Install)
	if err != nil {
		return nil, nil, err
	}
	deployed, err := s.deployedInstallComponentIDs(ctx, orgID, install.ID)
	if err != nil {
		return nil, nil, err
	}
	deployedIDs := make([]string, 0, len(deployed))
	for id, active := range deployed {
		if active {
			deployedIDs = append(deployedIDs, id)
		}
	}

	query := s.chDB.WithContext(ctx).
		Scopes(scopes.WithOverrideTable(app.InstallComponentResourceStatesLatestView)).
		Where(app.InstallComponentResourceState{
			OrgID:     orgID,
			InstallID: install.ID,
		}).
		Where(app.LatestReportOnlySQL(), app.LatestReportOnlyArgs(orgID, install.ID)...)
	if len(deployedIDs) == 0 {
		query = query.Where("source != ?", app.InstallComponentResourceSourceComponent)
	} else {
		query = query.Where(
			s.chDB.Where("source != ?", app.InstallComponentResourceSourceComponent).
				Or("install_component_id IN ?", deployedIDs),
		)
	}
	if in.InstallComponentID != "" {
		query = query.Where(app.InstallComponentResourceState{InstallComponentID: in.InstallComponentID})
	}
	if in.Kind != "" {
		query = query.Where(app.InstallComponentResourceState{Kind: in.Kind})
	}
	if in.Namespace != "" {
		query = query.Where(app.InstallComponentResourceState{Namespace: in.Namespace})
	}
	if in.Health != "" {
		query = query.Where(app.InstallComponentResourceState{Health: in.Health})
	}
	if in.Provider != "" {
		query = query.Where(app.InstallComponentResourceState{Provider: in.Provider})
	}

	var resources []app.InstallComponentResourceState
	if err := query.
		Order("source, install_component_id, owner_name, provider, kind, namespace, name").
		Limit(limit + 1).
		Offset(offset).
		Find(&resources).Error; err != nil {
		return nil, nil, fmt.Errorf("unable to list install resources: %w", err)
	}
	resources, hasMore := apiPkg.MCPClipList(resources, limit)

	declared, err := s.declaredProbesByInstallComponent(ctx, orgID, install.ID)
	if err != nil {
		return nil, nil, err
	}
	markRemovedProbes(resources, declared)

	componentNames := make(map[string]string, len(install.InstallComponents))
	for _, component := range install.InstallComponents {
		componentNames[component.ID] = component.Component.Name
	}
	out := make([]mcpInstallResource, 0, len(resources))
	for _, resource := range resources {
		out = append(out, mcpInstallResource{
			InstallComponentID: resource.InstallComponentID,
			ComponentID:        resource.ComponentID,
			ComponentName:      componentNames[resource.InstallComponentID],
			Source:             resource.Source,
			OwnerName:          resource.OwnerName,
			Provider:           resource.Provider,
			APIGroup:           resource.APIGroup,
			Kind:               resource.Kind,
			Namespace:          resource.Namespace,
			Name:               resource.Name,
			Health:             resource.Health,
			Message:            resource.Message,
			NativeStatus:       resource.NativeStatus,
			ObservedAt:         apiPkg.MCPTime(resource.ObservedAt),
			RemovedFromConfig:  resource.RemovedFromConfig,
		})
	}

	return apiPkg.MCPJSONResult(mcpListInstallResourcesResult{
		InstallID:   install.ID,
		InstallName: install.Name,
		Resources:   out,
		HasMore:     hasMore,
		Limit:       limit,
		Offset:      offset,
		NextOffset:  apiPkg.MCPNextOffset(offset, limit, hasMore),
		NextActions: []mcpNextAction{
			{
				Action: "back",
				Label:  "Back to install summary",
				Tool:   "install_summary",
				Arguments: map[string]any{
					"install": install.ID,
				},
			},
			{
				Action: "stop",
				Label:  "Do nothing",
			},
		},
	})
}
