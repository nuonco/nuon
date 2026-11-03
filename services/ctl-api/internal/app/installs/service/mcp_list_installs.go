package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpListInstallsInput struct {
	AppID  string `json:"app_id,omitempty" jsonschema:"filter installs by app ID"`
	Limit  int    `json:"limit,omitempty" jsonschema:"maximum installs to return (default 20, max 100)"`
	Offset int    `json:"offset,omitempty" jsonschema:"skip this many installs; use next_offset from a previous response when has_more is true"`
}

type mcpInstallListItem struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	AppID           string `json:"app_id,omitempty"`
	AppName         string `json:"app_name,omitempty"`
	SandboxStatus   string `json:"sandbox_status,omitempty"`
	ComponentStatus string `json:"component_status,omitempty"`
	RunnerStatus    string `json:"runner_status,omitempty"`
	CloudPlatform   string `json:"cloud_platform,omitempty"`
	CreatedAt       string `json:"created_at"`
}

type mcpListInstallsResult struct {
	Installs   []mcpInstallListItem `json:"installs"`
	HasMore    bool                 `json:"has_more"`
	Limit      int                  `json:"limit"`
	Offset     int                  `json:"offset"`
	NextOffset int                  `json:"next_offset,omitempty"`
}

func (s *service) mcpListInstalls(ctx context.Context, _ *mcp.CallToolRequest, in mcpListInstallsInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Org(ctx)
	if err != nil {
		return nil, nil, err
	}

	limit, offset, err := apiPkg.MCPListPage(in.Limit, in.Offset)
	if err != nil {
		return nil, nil, err
	}

	var installs []app.Install
	err = authz.New(authz.Params{DB: s.db}).ReadCustomerResources(ctx, orgID, func(tx, _, query *gorm.DB) error {
		query = query.
			Preload("App", func(db *gorm.DB) *gorm.DB { return db.Select("id", "name") }).
			Preload("AppRunnerConfig").
			Preload("AppConfig", func(db *gorm.DB) *gorm.DB {
				return db.Select("id", "app_id", "app_branch_id")
			}).
			Preload("AppConfig.RunnerConfig").
			Preload("RunnerGroup").
			Preload("RunnerGroup.Runners").
			Where(app.Install{OrgID: orgID}).
			Order("name ASC, id ASC").
			Limit(limit + 1).
			Offset(offset)

		if in.AppID != "" {
			query = query.Where(app.Install{AppID: in.AppID})
		}

		return query.Find(&installs).Error
	})
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list installs: %w", err)
	}

	installs, hasMore := apiPkg.MCPClipList(installs, limit)

	out := make([]mcpInstallListItem, 0, len(installs))
	for _, inst := range installs {
		item := mcpInstallListItem{
			ID:              inst.ID,
			Name:            inst.Name,
			AppID:           inst.AppID,
			AppName:         inst.App.Name,
			SandboxStatus:   string(inst.SandboxStatus),
			ComponentStatus: string(inst.CompositeComponentStatus),
			RunnerStatus:    string(inst.RunnerStatus),
			CloudPlatform:   string(inst.CloudPlatform),
			CreatedAt:       apiPkg.MCPTime(inst.CreatedAt),
		}
		out = append(out, item)
	}

	return apiPkg.MCPJSONResult(mcpListInstallsResult{
		Installs:   out,
		HasMore:    hasMore,
		Limit:      limit,
		Offset:     offset,
		NextOffset: apiPkg.MCPNextOffset(offset, limit, hasMore),
	})
}
