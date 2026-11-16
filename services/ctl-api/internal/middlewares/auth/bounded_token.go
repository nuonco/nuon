package auth

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (m *middleware) authenticateBoundedToken(ctx *gin.Context, token *app.Token) (*app.Account, error) {
	api, declared := require.OperationAPI(ctx)
	if !declared || api != permissions.APICtl || cctx.IsGlobal(ctx) {
		return nil, fmt.Errorf("bounded tokens require an org-scoped operation-based ctl-api route")
	}
	if token.TokenType != app.TokenTypeStatic || token.SourceType != "" || token.SourceID != "" {
		return nil, fmt.Errorf("bounded-token authentication currently supports direct static service-account tokens only")
	}
	account, err := m.acctClient.FetchAccountIdentity(ctx, token.AccountID)
	if err != nil {
		return nil, err
	}
	if account.AccountType != app.AccountTypeService {
		return nil, fmt.Errorf("bounded tokens require a service account")
	}
	catalog, err := permissions.DefaultCatalog()
	if err != nil {
		return nil, err
	}
	authority, err := m.authzClient.LoadAuthority(ctx, catalog, authz.AuthorityInput{
		AccountID: account.ID, CredentialID: token.ID, OrgID: token.OrgID, API: api, Ceiling: token.PermissionCeiling,
	})
	if err != nil {
		return nil, err
	}
	authz.SetAuthorityGinContext(ctx, authority)
	return account, nil
}
