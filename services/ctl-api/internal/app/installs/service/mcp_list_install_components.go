package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type mcpListInstallComponentsInput struct {
	Install string `json:"install" jsonschema:"install name or ID"`
}

type mcpInstallComponentItem struct {
	ID                      string                  `json:"id"`
	ComponentID             string                  `json:"component_id"`
	Name                    string                  `json:"name"`
	Type                    string                  `json:"type"`
	Status                  string                  `json:"status,omitempty"`
	StatusDescription       string                  `json:"status_description,omitempty"`
	HealthStatus            string                  `json:"health_status,omitempty"`
	HealthStatusDescription string                  `json:"health_status_description,omitempty"`
	LatestDeploy            *mcpInstallLatestDeploy `json:"latest_deploy,omitempty"`
}

type mcpInstallLatestDeploy struct {
	ID                string `json:"id"`
	BuildID           string `json:"build_id,omitempty"`
	Status            string `json:"status"`
	StatusDescription string `json:"status_description,omitempty"`
	CreatedAt         string `json:"created_at"`
}

func (s *service) mcpListInstallComponents(ctx context.Context, _ *mcp.CallToolRequest, in mcpListInstallComponentsInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.Install == "" {
		return nil, nil, fmt.Errorf("install is required")
	}

	var install app.Install
	err = s.db.WithContext(ctx).
		Where(app.Install{OrgID: orgID}).
		Where(s.db.Where(app.Install{ID: in.Install}).Or(app.Install{Name: in.Install})).
		First(&install).Error
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find install %q: %w", in.Install, err)
	}

	var components []app.InstallComponent
	err = s.db.WithContext(ctx).
		Preload("Component").
		Preload("InstallDeploys", func(db *gorm.DB) *gorm.DB {
			return db.
				Scopes(scopes.WithOverrideTable(views.CustomViewName(s.db, &app.InstallDeploy{}, "latest_view_v1"))).
				Order("install_deploys_latest_view_v1.created_at DESC")
		}).
		Where(app.InstallComponent{InstallID: install.ID}).
		Find(&components).Error
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list install components: %w", err)
	}

	sort.Slice(components, func(i, j int) bool {
		return components[i].Component.Name < components[j].Component.Name
	})
	result := make([]mcpInstallComponentItem, 0, len(components))
	for _, component := range components {
		item := mcpInstallComponentItem{
			ID:                      component.ID,
			ComponentID:             component.ComponentID,
			Name:                    component.Component.Name,
			Type:                    string(component.Component.Type),
			Status:                  string(component.Status),
			StatusDescription:       component.StatusDescription,
			HealthStatus:            string(component.HealthStatus),
			HealthStatusDescription: component.HealthStatusDescription,
		}
		if len(component.InstallDeploys) > 0 {
			deploy := component.InstallDeploys[0]
			item.LatestDeploy = &mcpInstallLatestDeploy{
				ID:                deploy.ID,
				BuildID:           deploy.ComponentBuildID,
				Status:            string(deploy.Status),
				StatusDescription: deploy.StatusDescription,
				CreatedAt:         apiPkg.MCPTime(deploy.CreatedAt),
			}
		}
		result = append(result, item)
	}

	return apiPkg.MCPJSONResult(result)
}
