package service

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type mcpWhoamiInput struct{}

func (s *service) mcpWhoami(ctx context.Context, _ *mcp.CallToolRequest, _ mcpWhoamiInput) (*mcp.CallToolResult, any, error) {
	acct, err := cctx.AccountFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	orgs := make([]map[string]string, 0, len(acct.Orgs))
	for _, org := range acct.Orgs {
		orgs = append(orgs, map[string]string{"id": org.ID, "name": org.Name})
	}

	out := map[string]any{
		"account": map[string]any{
			"id":       acct.ID,
			"email":    acct.Email,
			"employee": acct.IsEmployee,
		},
		"orgs":        orgs,
		"current_org": nil,
	}

	if orgID := keys.OrgIDFromContext(ctx); orgID != "" {
		currentOrg := map[string]string{"id": orgID}
		for _, org := range acct.Orgs {
			if org.ID == orgID {
				currentOrg["name"] = org.Name
				break
			}
		}
		out["current_org"] = currentOrg
	} else if len(orgs) > 1 {
		out["hint"] = "No org selected. Run nuon orgs select."
	}

	return api.MCPJSONResult(out)
}
