package service

import (
	"context"
	"slices"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type mcpWhoamiInput struct{}

func (s *service) mcpWhoami(ctx context.Context, _ *mcp.CallToolRequest, _ mcpWhoamiInput) (*mcp.CallToolResult, any, error) {
	out := map[string]any{"current_org": nil}
	err := authz.New(authz.Params{DB: s.db}).ReadCustomerNavigation(ctx, func(tx *gorm.DB, orgIDs []string) error {
		credential, _ := cctx.VerifiedCredentialFromContext(ctx)
		var acct app.Account
		if err := tx.Select("id", "email").Where(app.Account{ID: credential.AccountID}).Take(&acct).Error; err != nil {
			return err
		}
		out["account"] = map[string]string{"id": acct.ID, "email": acct.Email}
		out["org_count"] = len(orgIDs)
		if orgID := keys.OrgIDFromContext(ctx); slices.Contains(orgIDs, orgID) {
			var org OrgListItem
			if err := tx.Model(&app.Org{}).Select("id", "name").Where(app.Org{ID: orgID}).Take(&org).Error; err != nil {
				return err
			}
			out["current_org"] = org
		} else {
			out["hint"] = "No accessible org on this request. Use list_orgs and send X-Nuon-Org-ID."
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return apiPkg.MCPJSONResult(out)
}
