package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type mcpSelectOrgInput struct {
	OrgID string `json:"org_id" jsonschema:"the ID of the org to make active for subsequent tool calls"`
}

func (s *service) mcpSelectOrg(ctx context.Context, _ *mcp.CallToolRequest, in mcpSelectOrgInput) (*mcp.CallToolResult, any, error) {
	if in.OrgID == "" {
		return nil, nil, fmt.Errorf("org_id is required")
	}

	selector := keys.OrgSelectorFromContext(ctx)
	if selector == nil {
		return nil, nil, fmt.Errorf("org selection is not available in this context")
	}

	accountID := keys.CreatedByIDFromContext(ctx)

	var acct app.Account
	if res := s.db.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Org").
		First(&acct, "id = ?", accountID); res.Error != nil {
		return nil, nil, fmt.Errorf("unable to fetch account: %w", res.Error)
	}

	var selected *app.Org
	for _, o := range acct.Orgs {
		if o.ID == in.OrgID {
			selected = o
			break
		}
	}
	if selected == nil {
		return nil, nil, fmt.Errorf("account does not have access to org %s", in.OrgID)
	}

	selector(selected.ID)

	return apiPkg.MCPJSONResult(map[string]any{
		"selected_org": map[string]string{"id": selected.ID, "name": selected.Name},
		"note":         "Active org set. It applies to subsequent tool calls made with this token.",
	})
}
