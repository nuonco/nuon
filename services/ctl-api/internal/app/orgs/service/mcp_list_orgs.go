package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type mcpListOrgsInput struct{}

func (s *service) mcpListOrgs(ctx context.Context, _ *mcp.CallToolRequest, _ mcpListOrgsInput) (*mcp.CallToolResult, any, error) {
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

	return apiPkg.MCPJSONResult(map[string]any{
		"orgs":           orgs,
		"current_org_id": keys.OrgIDFromContext(ctx),
	})
}
