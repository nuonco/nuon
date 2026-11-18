package auth

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (m *middleware) authenticateBoundedToken(ctx *gin.Context, token *app.Token) (*app.Account, error) {
	api, declared := require.OperationAPI(ctx)
	if !declared || cctx.IsGlobal(ctx) {
		return nil, fmt.Errorf("bounded tokens require an org-scoped operation-based route")
	}
	direct := token.TokenType == app.TokenTypeStatic && token.SourceType == "" && token.SourceID == ""
	jobToken := token.TokenType == app.TokenTypeNuon && token.SourceType == app.TokenSourceTypeRunnerJob && token.SourceID != ""
	if !direct && !jobToken {
		return nil, fmt.Errorf("bounded-token authentication requires a direct static or runner-job credential")
	}
	account, err := m.fetchGrantAccount(ctx, token.AccountID, api)
	if err != nil {
		return nil, err
	}
	if account.AccountType != app.AccountTypeService {
		return nil, fmt.Errorf("bounded tokens require a service account")
	}
	if jobToken {
		if account.Subject != token.SourceID || token.OrgID == "" {
			return nil, fmt.Errorf("job token must identify its job service account and org")
		}
		var job app.RunnerJob
		if err := m.db.WithContext(ctx.Request.Context()).Select("id").Where(app.RunnerJob{ID: token.SourceID, OrgID: token.OrgID}).Take(&job).Error; err != nil {
			return nil, fmt.Errorf("job token requires a live job in its org: %w", err)
		}
	}
	catalog, err := permissions.DefaultCatalog()
	if err != nil {
		return nil, err
	}
	authority, err := m.authzClient.LoadAuthority(ctx, catalog, authz.AuthorityInput{
		AccountID: account.ID, CredentialID: token.ID, OrgID: token.OrgID, API: api, Ceiling: token.PermissionCeiling,
		CredentialOrgID: token.OrgID,
	})
	if err != nil {
		return nil, err
	}
	authz.SetAuthorityGinContext(ctx, authority)
	return account, nil
}

func (m *middleware) authenticateAccount(ctx *gin.Context, token *app.Token) (*app.Account, error) {
	api, declared := require.OperationAPI(ctx)
	if cctx.IsGlobal(ctx) {
		return m.acctClient.FetchAccount(ctx, token.AccountID)
	}
	orgID := ctx.GetHeader("X-Nuon-Org-ID")
	if orgID == "" {
		orgID = ctx.Query("org_id")
	}
	if orgID == "" && declared && api != permissions.APICtl {
		account, err := m.acctClient.FetchAccount(ctx, token.AccountID)
		if err != nil || len(account.OrgIDs) != 1 {
			return account, err
		}
		orgID = account.OrgIDs[0]
	}
	if orgID == "" {
		return m.acctClient.FetchAccount(ctx, token.AccountID)
	}
	usesGrants, err := m.authzClient.AccountUsesGrants(ctx, token.AccountID, orgID)
	if err != nil {
		return nil, err
	}
	if !usesGrants {
		return m.acctClient.FetchAccount(ctx, token.AccountID)
	}
	if !declared || (token.OrgID != "" && token.OrgID != orgID) {
		return nil, stderr.ErrAuthorization{Err: fmt.Errorf("token is not allowed in this org or API"), Description: "token is not allowed in this org or API"}
	}
	account, err := m.fetchGrantAccount(ctx, token.AccountID, api)
	if err != nil {
		return nil, err
	}
	catalog, err := permissions.DefaultCatalog()
	if err != nil {
		return nil, err
	}
	authority, err := m.authzClient.LoadAuthority(ctx, catalog, authz.AuthorityInput{
		AccountID: account.ID, CredentialID: token.ID, OrgID: orgID, API: api,
		CredentialOrgID: token.OrgID,
	})
	if err != nil {
		m.l.Warn("grant authentication denied", zap.Error(err))
		return nil, stderr.ErrAuthorization{Err: fmt.Errorf("unable to authorize account grants"), Description: "unable to authorize account grants"}
	}
	authz.SetAuthorityGinContext(ctx, authority)
	return account, nil
}

func (m *middleware) fetchGrantAccount(ctx *gin.Context, accountID string, api permissions.API) (*app.Account, error) {
	if api == permissions.APIRunner || api == permissions.APIStack {
		return m.acctClient.FetchAccount(ctx, accountID)
	}
	return m.acctClient.FetchAccountIdentity(ctx, accountID)
}
