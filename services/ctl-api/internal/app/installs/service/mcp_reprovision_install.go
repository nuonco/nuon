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
	RequestID string `json:"request_id,omitempty" jsonschema:"optional idempotency key. The same id and body returns the original workflow. A different body, or an install that has moved to another app config, is rejected"`
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
	operation := "reprovision"
	var hashed any = ReprovisionInstallRequest{PlanOnly: in.PlanOnly, Role: in.Role}
	if in.StackOnly {
		workflowType = app.WorkflowTypeReprovisionStack
		operation = "reprovision-stack"
		hashed = struct {
			PlanOnly  bool   `json:"plan_only"`
			Role      string `json:"role"`
			StackOnly bool   `json:"stack_only"`
		}{PlanOnly: in.PlanOnly, Role: in.Role, StackOnly: true}
	}
	requestID, requestHash, operation, err := mcpInstallRequest(in.RequestID, operation, hashed)
	if err != nil {
		return nil, nil, err
	}

	started, err := s.startInstallWorkflowWithRequestID(ctx, orgID, in.Install, workflowType, in.PlanOnly, in.Role, nil, requestID, requestHash, operation)
	if err != nil {
		return nil, nil, err
	}
	return apiPkg.MCPJSONResult(started)
}
