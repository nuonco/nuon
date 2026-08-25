package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type mcpWhoamiInput struct{}

func (s *service) mcpWhoami(ctx context.Context, _ *mcp.CallToolRequest, _ mcpWhoamiInput) (*mcp.CallToolResult, any, error) {
	accountID := keys.CreatedByIDFromContext(ctx)

	var acct app.Account
	if res := s.db.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Org").
		First(&acct, "id = ?", accountID); res.Error != nil {
		return nil, nil, fmt.Errorf("unable to fetch account: %w", res.Error)
	}

	orgs := make([]map[string]string, 0, len(acct.Orgs))
	for _, o := range acct.Orgs {
		orgs = append(orgs, map[string]string{"id": o.ID, "name": o.Name})
	}

	out := map[string]any{
		"account":     map[string]string{"id": acct.ID, "email": acct.Email},
		"orgs":        orgs,
		"current_org": nil,
	}

	if orgID := keys.OrgIDFromContext(ctx); orgID != "" {
		org, err := s.getOrg(ctx, orgID)
		if err != nil {
			return nil, nil, err
		}
		out["current_org"] = map[string]string{"id": org.ID, "name": org.Name}
	} else if len(orgs) > 1 {
		out["hint"] = "No org selected. Call select_org with one of the org IDs above."
	}

	return apiPkg.MCPJSONResult(out)
}
