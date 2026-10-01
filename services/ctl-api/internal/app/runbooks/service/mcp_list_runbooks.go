package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type mcpListRunbooksInput struct {
	AppID   string `json:"app_id,omitempty" jsonschema:"app ID; required when install is omitted"`
	Install string `json:"install,omitempty" jsonschema:"install name or ID; lists runbooks pinned to this install's app config"`
	Synced  *bool  `json:"synced,omitempty" jsonschema:"when install is set, true (default) lists runbooks in the current app config; false lists runbooks no longer in it"`
	Q       string `json:"q,omitempty" jsonschema:"search by runbook name or ID"`
	Limit   int    `json:"limit,omitempty" jsonschema:"maximum runbooks to return (default 20, max 100)"`
	Offset  int    `json:"offset,omitempty" jsonschema:"skip this many runbooks; use next_offset from a previous response when has_more is true"`
}

type mcpRunbookListItem struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	Status             string `json:"status,omitempty"`
	CreatedAt          string `json:"created_at"`
	InstallRunbookID   string `json:"install_runbook_id,omitempty"`
	RunbookConfigID    string `json:"runbook_config_id,omitempty"`
	StepCount          int    `json:"step_count,omitempty"`
	InCurrentAppConfig *bool  `json:"in_current_app_config,omitempty"`
	Runnable           *bool  `json:"runnable,omitempty"`
}

type mcpListRunbooksResult struct {
	Runbooks   []mcpRunbookListItem `json:"runbooks"`
	HasMore    bool                 `json:"has_more"`
	Limit      int                  `json:"limit"`
	Offset     int                  `json:"offset"`
	NextOffset int                  `json:"next_offset,omitempty"`
}

func (s *service) mcpListRunbooks(ctx context.Context, _ *mcp.CallToolRequest, in mcpListRunbooksInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.AppID == "" && in.Install == "" {
		return nil, nil, fmt.Errorf("install or app_id is required")
	}

	limit, offset, err := apiPkg.MCPListPage(in.Limit, in.Offset)
	if err != nil {
		return nil, nil, err
	}

	var (
		items   []mcpRunbookListItem
		hasMore bool
	)
	if in.Install != "" {
		items, hasMore, err = s.listInstallRunbooks(ctx, orgID, in, limit, offset)
	} else {
		items, hasMore, err = s.listAppRunbooks(ctx, orgID, in.AppID, limit, offset)
	}
	if err != nil {
		return nil, nil, err
	}

	return apiPkg.MCPJSONResult(mcpListRunbooksResult{
		Runbooks:   items,
		HasMore:    hasMore,
		Limit:      limit,
		Offset:     offset,
		NextOffset: apiPkg.MCPNextOffset(offset, limit, hasMore),
	})
}

func (s *service) listAppRunbooks(ctx context.Context, orgID, appID string, limit, offset int) ([]mcpRunbookListItem, bool, error) {
	var runbooks []app.Runbook
	err := s.db.WithContext(ctx).
		Where(app.Runbook{OrgID: orgID, AppID: appID}).
		Order("created_at DESC").
		Limit(limit + 1).
		Offset(offset).
		Find(&runbooks).Error
	if err != nil {
		return nil, false, fmt.Errorf("unable to list runbooks: %w", err)
	}

	runbooks, hasMore := apiPkg.MCPClipList(runbooks, limit)
	out := make([]mcpRunbookListItem, 0, len(runbooks))
	for _, r := range runbooks {
		out = append(out, mcpRunbookListItem{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Status:      string(r.Status),
			CreatedAt:   apiPkg.MCPTime(r.CreatedAt),
		})
	}
	return out, hasMore, nil
}

func (s *service) listInstallRunbooks(ctx context.Context, orgID string, in mcpListRunbooksInput, limit, offset int) ([]mcpRunbookListItem, bool, error) {
	install, err := s.resolveInstallRef(ctx, orgID, in.Install)
	if err != nil {
		return nil, false, err
	}
	if in.AppID != "" && install.AppID != in.AppID {
		return nil, false, fmt.Errorf("install %q belongs to a different app", in.Install)
	}

	synced := true
	if in.Synced != nil {
		synced = *in.Synced
	}

	var installRunbooks []app.InstallRunbook
	tx := s.db.WithContext(ctx).
		Joins("JOIN runbooks ON runbooks.id = install_runbooks.runbook_id AND runbooks.deleted_at = 0").
		Preload("Runbook").
		Preload("Runbook.Configs", func(tx *gorm.DB) *gorm.DB {
			if install.AppConfigID == "" {
				return tx.Scopes(scopes.WithOverrideTable("runbook_configs_latest_view_v1"))
			}
			return tx.Where(app.RunbookConfig{AppConfigID: install.AppConfigID})
		}).
		Preload("Runbook.Configs.Steps", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("idx ASC")
		}).
		Where(app.InstallRunbook{OrgID: orgID, InstallID: install.ID}).
		Where(appConfigRunbookFilter(synced)).
		Order("install_runbooks.created_at DESC").
		Limit(limit + 1).
		Offset(offset)
	if in.Q != "" {
		tx = tx.Where("runbooks.name ILIKE ? OR install_runbooks.runbook_id = ? OR install_runbooks.id = ?", "%"+in.Q+"%", in.Q, in.Q)
	}
	if err := tx.Find(&installRunbooks).Error; err != nil {
		return nil, false, fmt.Errorf("unable to list install runbooks: %w", err)
	}

	installRunbooks, hasMore := apiPkg.MCPClipList(installRunbooks, limit)
	out := make([]mcpRunbookListItem, 0, len(installRunbooks))
	for _, installRunbook := range installRunbooks {
		item := mcpRunbookListItem{
			ID:               installRunbook.RunbookID,
			Name:             installRunbook.Runbook.Name,
			Description:      installRunbook.Runbook.Description,
			Status:           string(installRunbook.Runbook.Status),
			CreatedAt:        apiPkg.MCPTime(installRunbook.CreatedAt),
			InstallRunbookID: installRunbook.ID,
		}
		if installRunbook.RunbookID == "" {
			item.ID = installRunbook.Runbook.ID
		}
		inConfig := len(installRunbook.Runbook.Configs) > 0
		item.InCurrentAppConfig = &inConfig
		item.Runnable = &inConfig
		if inConfig {
			cfg := installRunbook.Runbook.Configs[0]
			item.RunbookConfigID = cfg.ID
			item.StepCount = len(cfg.Steps)
		}
		out = append(out, item)
	}
	return out, hasMore, nil
}
