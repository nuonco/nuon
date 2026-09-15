package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpReprovisionInstallInput struct {
	Install   string `json:"install" jsonschema:"install name or ID"`
	PlanOnly  bool   `json:"plan_only,omitempty" jsonschema:"if true, only plan the reprovision; do not apply"`
	StackOnly bool   `json:"stack_only,omitempty" jsonschema:"if true, only reprovision the install stack (runner infra); sandbox and components are left unchanged"`
	Role      string `json:"role,omitempty" jsonschema:"optional IAM role name from list_available_roles; omit to use the default"`
}

func (s *service) mcpReprovisionInstall(ctx context.Context, _ *mcp.CallToolRequest, in mcpReprovisionInstallInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Write(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.Install == "" {
		return nil, nil, fmt.Errorf("install is required")
	}

	workflowType := app.WorkflowTypeReprovision
	if in.StackOnly {
		workflowType = app.WorkflowTypeReprovisionStack
	}

	started, err := s.startInstallWorkflow(ctx, orgID, in.Install, workflowType, in.PlanOnly, in.Role, nil)
	if err != nil {
		return nil, nil, err
	}
	return apiPkg.MCPJSONResult(started)
}
