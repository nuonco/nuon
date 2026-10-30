package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpListScheduledReconcilesInput struct {
	Install string `json:"install" jsonschema:"install name or ID"`
	Status  string `json:"status,omitempty" jsonschema:"filter by status: pending, fired, no-changes, cancelled, failed"`
	Limit   int    `json:"limit,omitempty" jsonschema:"maximum scheduled reconciles to return (default 20, max 100)"`
	Offset  int    `json:"offset,omitempty" jsonschema:"skip this many scheduled reconciles; use next_offset from a previous response when has_more is true"`
}

type mcpScheduledReconcileItem struct {
	ID           string `json:"id"`
	ScheduledAt  string `json:"scheduled_at"`
	PlanOnly     bool   `json:"plan_only"`
	Status       string `json:"status"`
	StatusReason string `json:"status_reason,omitempty"`
	WorkflowID   string `json:"workflow_id,omitempty"`
	CreatedAt    string `json:"created_at"`
}

type mcpListScheduledReconcilesResult struct {
	ScheduledReconciles []mcpScheduledReconcileItem `json:"scheduled_reconciles"`
	HasMore             bool                        `json:"has_more"`
	Limit               int                         `json:"limit"`
	Offset              int                         `json:"offset"`
	NextOffset          int                         `json:"next_offset,omitempty"`
}

func (s *service) mcpListScheduledReconciles(ctx context.Context, _ *mcp.CallToolRequest, in mcpListScheduledReconcilesInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.Install == "" {
		return nil, nil, fmt.Errorf("install is required")
	}

	install, err := s.findInstall(ctx, orgID, in.Install)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get install %q: %w", in.Install, err)
	}

	limit, offset, err := apiPkg.MCPListPage(in.Limit, in.Offset)
	if err != nil {
		return nil, nil, err
	}

	tx := s.db.WithContext(ctx).
		Where(app.InstallScheduledReconcile{InstallID: install.ID, OrgID: orgID}).
		Order("created_at DESC").
		Limit(limit + 1).
		Offset(offset)
	if in.Status != "" {
		tx = tx.Where(app.InstallScheduledReconcile{Status: app.InstallScheduledReconcileStatus(in.Status)})
	}

	var schedules []app.InstallScheduledReconcile
	if res := tx.Find(&schedules); res.Error != nil {
		return nil, nil, fmt.Errorf("unable to list scheduled reconciles: %w", res.Error)
	}

	schedules, hasMore := apiPkg.MCPClipList(schedules, limit)

	out := make([]mcpScheduledReconcileItem, 0, len(schedules))
	for _, schedule := range schedules {
		item := mcpScheduledReconcileItem{
			ID:           schedule.ID,
			ScheduledAt:  apiPkg.MCPTime(schedule.ScheduledAt),
			PlanOnly:     schedule.PlanOnly,
			Status:       string(schedule.Status),
			StatusReason: schedule.StatusReason,
			CreatedAt:    apiPkg.MCPTime(schedule.CreatedAt),
		}
		if schedule.WorkflowID != nil {
			item.WorkflowID = *schedule.WorkflowID
		}
		out = append(out, item)
	}

	return apiPkg.MCPJSONResult(mcpListScheduledReconcilesResult{
		ScheduledReconciles: out,
		HasMore:             hasMore,
		Limit:               limit,
		Offset:              offset,
		NextOffset:          apiPkg.MCPNextOffset(offset, limit, hasMore),
	})
}
