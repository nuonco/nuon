package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpCancelScheduledReconcileInput struct {
	Install              string `json:"install" jsonschema:"install name or ID"`
	ScheduledReconcileID string `json:"scheduled_reconcile_id" jsonschema:"ID of the scheduled reconcile to cancel"`
}

type mcpCancelScheduledReconcileResult struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func (s *service) mcpCancelScheduledReconcile(ctx context.Context, _ *mcp.CallToolRequest, in mcpCancelScheduledReconcileInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Write(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.Install == "" {
		return nil, nil, fmt.Errorf("install is required")
	}
	if in.ScheduledReconcileID == "" {
		return nil, nil, fmt.Errorf("scheduled_reconcile_id is required")
	}

	install, err := s.findInstall(ctx, orgID, in.Install)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get install %q: %w", in.Install, err)
	}

	var schedule app.InstallScheduledReconcile
	if err := s.db.WithContext(ctx).
		Where(app.InstallScheduledReconcile{
			ID:        in.ScheduledReconcileID,
			InstallID: install.ID,
			OrgID:     orgID,
		}).
		First(&schedule).Error; err != nil {
		return nil, nil, fmt.Errorf("unable to find scheduled reconcile %q: %w", in.ScheduledReconcileID, err)
	}

	if err := s.helpers.CancelInstallScheduledReconcile(ctx, &schedule, "cancelled by user"); err != nil {
		if errors.Is(err, installhelpers.ErrScheduledReconcileNotPending) {
			return nil, nil, fmt.Errorf("the scheduled reconcile is not pending")
		}
		return nil, nil, err
	}

	return apiPkg.MCPJSONResult(mcpCancelScheduledReconcileResult{
		ID:     schedule.ID,
		Status: string(schedule.Status),
	})
}
