package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpDeprovisionSandboxInput struct {
	Install  string `json:"install" jsonschema:"install name or ID"`
	Confirm  bool   `json:"confirm" jsonschema:"must be true to apply deprovision (not required for plan_only)"`
	PlanOnly bool   `json:"plan_only,omitempty" jsonschema:"if true, only plan sandbox deprovision; do not apply"`
	Role     string `json:"role,omitempty" jsonschema:"optional custom/IAM role name for the workflow"`
}

func (s *service) mcpDeprovisionSandbox(ctx context.Context, _ *mcp.CallToolRequest, in mcpDeprovisionSandboxInput) (*mcp.CallToolResult, any, error) {
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

	started, err := s.startInstallWorkflow(ctx, orgID, in.Install, app.WorkflowTypeDeprovisionSandbox, in.PlanOnly, in.Role, nil)
	if err != nil {
		return nil, nil, err
	}
	return apiPkg.MCPJSONResult(started)
}
