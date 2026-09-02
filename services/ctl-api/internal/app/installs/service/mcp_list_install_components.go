package service

import (
	"context"
	"fmt"

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

func (s *service) mcpListInstallComponents(ctx context.Context, _ *mcp.CallToolRequest, in mcpListInstallComponentsInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	var install app.Install
	err = s.db.WithContext(ctx).
		Where(app.Install{OrgID: orgID}).
		Where("id = ? OR name = ?", in.Install, in.Install).
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

	return apiPkg.MCPJSONResult(components)
}
