package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpReprovisionSandboxInput struct {
	Install        string `json:"install" jsonschema:"install name or ID"`
	PlanOnly       bool   `json:"plan_only,omitempty" jsonschema:"if true, only plan sandbox reprovision; do not apply"`
	Role           string `json:"role,omitempty" jsonschema:"optional IAM role name from list_available_roles; omit to use the default"`
	SkipComponents bool   `json:"skip_components,omitempty" jsonschema:"if true, skip deploying components after sandbox reprovision"`
	RequestID      string `json:"request_id,omitempty" jsonschema:"optional idempotency key. The same id and body returns the original workflow. A different body, or an install that has moved to another app config, is rejected"`
}

func (s *service) mcpReprovisionSandbox(ctx context.Context, _ *mcp.CallToolRequest, in mcpReprovisionSandboxInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Write(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.Install == "" {
		return nil, nil, fmt.Errorf("install is required")
	}

	metadata := map[string]string{}
	if in.SkipComponents {
		metadata["skip_components"] = "true"
	}

	requestID, requestHash, operation, err := mcpInstallRequest(in.RequestID, "reprovision-sandbox", ReprovisionInstallSandboxRequest{
		Role:           in.Role,
		PlanOnly:       in.PlanOnly,
		SkipComponents: in.SkipComponents,
	})
	if err != nil {
		return nil, nil, err
	}
	started, err := s.startInstallWorkflowWithRequestID(ctx, orgID, in.Install, app.WorkflowTypeReprovisionSandbox, in.PlanOnly, in.Role, metadata, requestID, requestHash, operation)
	if err != nil {
		return nil, nil, err
	}
	return apiPkg.MCPJSONResult(started)
}
