package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/scheduledreconcile"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpReconcileInstallAppVersionInput struct {
	Install     string `json:"install" jsonschema:"install name or ID"`
	PlanOnly    bool   `json:"plan_only,omitempty" jsonschema:"if true, only plan the reconcile; do not apply"`
	ScheduledAt string `json:"scheduled_at,omitempty" jsonschema:"RFC3339 timestamp; if set, schedules the reconcile to fire at this time (must be 1 minute to 90 days from now) instead of running it immediately"`
}

type mcpReconcileInstallAppVersionResult struct {
	InstallID                 string         `json:"install_id"`
	InstallName               string         `json:"install_name"`
	WorkflowID                string         `json:"workflow_id,omitempty"`
	InstallAppConfigVersionID string         `json:"install_app_config_version_id,omitempty"`
	ScheduledReconcileID      string         `json:"scheduled_reconcile_id,omitempty"`
	ScheduledAt               string         `json:"scheduled_at,omitempty"`
	PlanOnly                  bool           `json:"plan_only"`
	NextAction                *mcpNextAction `json:"next_action,omitempty"`
}

func (s *service) mcpReconcileInstallAppVersion(ctx context.Context, _ *mcp.CallToolRequest, in mcpReconcileInstallAppVersionInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Write(ctx)
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

	if in.ScheduledAt != "" {
		scheduledAt, err := time.Parse(time.RFC3339, in.ScheduledAt)
		if err != nil {
			return nil, nil, fmt.Errorf("scheduled_at must be an RFC3339 timestamp: %w", err)
		}

		schedule, err := s.helpers.CreateInstallScheduledReconcileRow(ctx, install.ID, scheduledAt, in.PlanOnly)
		if err != nil {
			return nil, nil, scheduledReconcileMCPErr(err)
		}

		em, err := s.helpers.CreateScheduledReconcileEmitter(ctx, install.ID, schedule.ID, scheduledAt, scheduledreconcile.SignalType, &scheduledreconcile.Signal{
			InstallID:            install.ID,
			ScheduledReconcileID: schedule.ID,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("unable to create scheduled reconcile emitter: %w", err)
		}
		if err := s.helpers.SetInstallScheduledReconcileEmitterID(ctx, schedule.ID, em.ID); err != nil {
			return nil, nil, fmt.Errorf("unable to save scheduled reconcile emitter: %w", err)
		}

		return apiPkg.MCPJSONResult(mcpReconcileInstallAppVersionResult{
			InstallID:            install.ID,
			InstallName:          install.Name,
			ScheduledReconcileID: schedule.ID,
			ScheduledAt:          apiPkg.MCPTime(scheduledAt),
			PlanOnly:             in.PlanOnly,
		})
	}

	result, err := s.helpers.CreateReconcileAppVersionWorkflow(ctx, installhelpers.CreateReconcileAppVersionWorkflowInput{
		InstallID: install.ID,
		PlanOnly:  in.PlanOnly,
	})
	if err != nil {
		if errors.Is(err, installhelpers.ErrInstallAlreadyOnPinnedConfig) {
			return nil, nil, fmt.Errorf("the install is already on its pinned app config; there is nothing to reconcile")
		}
		if errors.Is(err, installhelpers.ErrInstallConfigUpdateInFlight) {
			return nil, nil, fmt.Errorf("a reconcile or app branch config update workflow is already queued or in progress for this install")
		}
		return nil, nil, err
	}

	return apiPkg.MCPJSONResult(mcpReconcileInstallAppVersionResult{
		InstallID:                 install.ID,
		InstallName:               install.Name,
		WorkflowID:                result.WorkflowID,
		InstallAppConfigVersionID: result.InstallAppConfigVersionID,
		PlanOnly:                  in.PlanOnly,
		NextAction:                mcpWatchStart(result.WorkflowID, ""),
	})
}

func scheduledReconcileMCPErr(err error) error {
	switch {
	case errors.Is(err, installhelpers.ErrScheduledReconcileAlreadyPending):
		return fmt.Errorf("the install already has a pending scheduled reconcile; cancel it before scheduling another")
	case errors.Is(err, installhelpers.ErrScheduledReconcileTimeTooSoon):
		return fmt.Errorf("scheduled_at must be at least 1 minute in the future")
	case errors.Is(err, installhelpers.ErrScheduledReconcileTimeTooFar):
		return fmt.Errorf("scheduled_at must be at most 90 days in the future")
	default:
		return err
	}
}
