package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpListAppBranchesInput struct {
	App string `json:"app" jsonschema:"app name or ID"`
}

type mcpListAppBranchItem struct {
	ID        string                   `json:"id"`
	Name      string                   `json:"name"`
	ManagedBy string                   `json:"managed_by"`
	LatestRun *mcpAppBranchRunListItem `json:"latest_run,omitempty"`
}

type mcpAppBranchRunListItem struct {
	ID               string `json:"id"`
	Status           string `json:"status"`
	Succeeded        bool   `json:"succeeded"`
	AwaitingApproval bool   `json:"awaiting_approval"`
	CreatedAt        string `json:"created_at"`
}

func (s *service) mcpListAppBranches(ctx context.Context, _ *mcp.CallToolRequest, in mcpListAppBranchesInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}
	if err := s.requireAppBranches(ctx); err != nil {
		return nil, nil, err
	}
	if in.App == "" {
		return nil, nil, fmt.Errorf("app is required")
	}

	a, err := s.findAppRef(ctx, orgID, in.App)
	if err != nil {
		return nil, nil, err
	}

	var branches []app.AppBranch
	res := s.db.WithContext(ctx).
		Where(app.AppBranch{OrgID: orgID, AppID: a.ID}).
		Order("name ASC").
		Find(&branches)
	if res.Error != nil {
		return nil, nil, fmt.Errorf("unable to list app branches: %w", res.Error)
	}

	out := make([]mcpListAppBranchItem, 0, len(branches))
	for _, b := range branches {
		item := mcpListAppBranchItem{ID: b.ID, Name: b.Name, ManagedBy: string(b.ManagedBy)}
		run, err := s.latestAppBranchRun(ctx, b.ID)
		if err != nil {
			return nil, nil, err
		}
		if run != nil {
			if err := s.markRunAwaitingApproval(ctx, run); err != nil {
				return nil, nil, err
			}
			item.LatestRun = &mcpAppBranchRunListItem{
				ID:               run.ID,
				Status:           run.Status,
				Succeeded:        run.Status == "success",
				AwaitingApproval: run.AwaitingApproval,
				CreatedAt:        run.CreatedAt.String(),
			}
		}
		out = append(out, item)
	}

	return apiPkg.MCPJSONResult(map[string]any{
		"app_id":   a.ID,
		"app_name": a.Name,
		"branches": out,
	})
}
