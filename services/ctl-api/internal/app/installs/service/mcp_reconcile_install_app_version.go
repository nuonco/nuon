package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpReconcileInstallAppVersionInput struct {
	Install  string `json:"install" jsonschema:"install name or ID"`
	PlanOnly bool   `json:"plan_only,omitempty" jsonschema:"if true, only plan the reconcile; do not apply"`
}

type mcpReconcileInstallAppVersionResult struct {
	InstallID                 string         `json:"install_id"`
	InstallName               string         `json:"install_name"`
	WorkflowID                string         `json:"workflow_id"`
	InstallAppConfigVersionID string         `json:"install_app_config_version_id"`
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
