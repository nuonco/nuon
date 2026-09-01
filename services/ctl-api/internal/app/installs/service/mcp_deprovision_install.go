package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/pkg/lifecyclephase"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpDeprovisionInstallInput struct {
	Install  string `json:"install" jsonschema:"install name or ID"`
	Confirm  bool   `json:"confirm" jsonschema:"must be true to apply deprovision (not required for plan_only)"`
	PlanOnly bool   `json:"plan_only,omitempty" jsonschema:"if true, only plan the deprovision; do not apply"`
}

func (s *service) mcpDeprovisionInstall(ctx context.Context, _ *mcp.CallToolRequest, in mcpDeprovisionInstallInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Write(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.Install == "" {
		return nil, nil, fmt.Errorf("install is required")
	}
	if err := mcpRequireDeprovisionConfirm(in.Confirm, in.PlanOnly); err != nil {
		return nil, nil, err
	}

	started, err := s.startInstallWorkflow(ctx, orgID, in.Install, app.WorkflowTypeDeprovision, in.PlanOnly, "", nil)
	if err != nil {
		return nil, nil, err
	}

	lp := lifecyclephase.New(lifecyclephase.Deprovisioning, "Tearing down components and cloud resources")
	if err := s.db.WithContext(ctx).Model(&app.Install{ID: started.InstallID}).Updates(map[string]any{
		"lifecycle_phase": lp,
	}).Error; err != nil {
		return nil, nil, fmt.Errorf("unable to update install lifecycle: %w", err)
	}

	return apiPkg.MCPJSONResult(started)
}
