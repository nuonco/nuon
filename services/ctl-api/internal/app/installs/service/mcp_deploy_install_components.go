package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpDeployInstallComponentsInput struct {
	Install  string `json:"install" jsonschema:"install name or ID"`
	PlanOnly bool   `json:"plan_only,omitempty" jsonschema:"if true, only plan component deploys; do not apply"`
	Role     string `json:"role,omitempty" jsonschema:"optional custom/IAM role name for component deploys"`
}

func (s *service) mcpDeployInstallComponents(ctx context.Context, _ *mcp.CallToolRequest, in mcpDeployInstallComponentsInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Write(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.Install == "" {
		return nil, nil, fmt.Errorf("install is required")
	}

	started, err := s.startInstallWorkflow(ctx, orgID, in.Install, app.WorkflowTypeDeployComponents, in.PlanOnly, in.Role, nil)
	if err != nil {
		return nil, nil, err
	}
	return apiPkg.MCPJSONResult(started)
}
