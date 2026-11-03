package service

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type CreateTokenRequest struct {
	AccountID string             `json:"account_id" binding:"required"`
	Role      string             `json:"role" binding:"required"`
	Scope     *permissions.Scope `json:"scope" binding:"required"`

	// Name labels the token where it is listed; defaults to the account's identity.
	Name string `json:"name"`

	// Duration defaults to one year.
	Duration string `json:"duration" default:"8760h"`

	Invalidate bool `json:"invalidate"`
}

type CreateTokenResponse struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	AccountID string            `json:"account_id"`
	Role      string            `json:"role"`
	Scope     permissions.Scope `json:"scope"`
	ExpiresAt time.Time         `json:"expires_at"`
	CreatedAt time.Time         `json:"created_at"`
	Token     string            `json:"token"`
}

// @ID						CreateToken
// @Summary				Create an API token for a service account in the current org
// @Description.markdown	create_token.md
// @Param					req	body	CreateTokenRequest	true	"Input"
// @Tags					accounts
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	CreateTokenResponse
// @Router					/v1/tokens [POST]
func (s *service) CreateToken(ctx *gin.Context) (require.PreparedMutation, error) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		return require.PreparedMutation{}, err
	}

	var req CreateTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return require.PreparedMutation{}, stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse request: %w", err),
			Description: fmt.Sprintf("unable to parse request: %s", err.Error()),
		}
	}

	scope, err := authz.RoleBindingScope(org.ID, req.Scope)
	if err != nil {
		return require.PreparedMutation{}, stderr.NewInvalidRequest(err)
	}
	if req.Duration == "" {
		req.Duration = accountshelpers.DefaultTokenDuration
	}
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		return require.PreparedMutation{}, stderr.NewInvalidRequest(fmt.Errorf("invalid duration: %w", err))
	}

	return require.PreparedMutation{SubjectAccountID: req.AccountID, Mutate: func(ctx *gin.Context) (require.Response, error) {
		access, err := require.MutationAccessFromContext(ctx)
		if err != nil {
			return require.Response{}, err
		}
		tx := access.DB()
		token, err := account.New(account.Params{DB: tx}).IssueCustomerToken(tx.Statement.Context, account.CustomerTokenRequest{
			AccountID: access.SubjectAccountID(), OrgID: access.Target().OrgID, Name: req.Name, Role: app.RoleType(req.Role), Scope: scope, Duration: duration, Invalidate: req.Invalidate,
		})
		if err != nil {
			return require.Response{}, err
		}
		return require.Response{Status: http.StatusCreated, Body: CreateTokenResponse{
			ID: token.ID, Name: token.Name, AccountID: token.AccountID, Role: token.Role, Scope: scope,
			ExpiresAt: token.ExpiresAt, CreatedAt: token.CreatedAt, Token: token.Token,
		}}, nil
	}}, nil
}

type TokenResponse struct {
	app.Token

	AccountName         string             `json:"account_name"`
	AccountEmail        string             `json:"account_email"`
	AccountRole         string             `json:"account_role"`
	AccountRoleBindings []app.RoleBinding  `json:"account_role_bindings"`
	Scope               *permissions.Scope `json:"scope,omitempty"`
}

// @ID						ListTokens
// @Summary				List API tokens for service accounts in the current org
// @Description.markdown	list_tokens.md
// @Param					account_id	query	string	false	"only list tokens for this service account"
// @Param					q			query	string	false	"case-insensitive substring match on token name, or exact match on token ID"
// @Param					offset		query	int		false	"offset of results to return"	Default(0)
// @Param					limit		query	int		false	"limit of results to return"	Default(10)
// @Param					page		query	int		false	"page number of results to return"	Default(0)
// @Tags					accounts
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		TokenResponse
// @Router					/v1/tokens [GET]
func (s *service) ListTokens(ctx *gin.Context) (any, error) {
	access, err := require.ReadAccessFromContext(ctx)
	if err != nil {
		return nil, err
	}
	tx := access.DB()
	requestCtx := tx.Statement.Context
	orgID := access.OrgID()
	helpers := accountshelpers.New(accountshelpers.Params{DB: tx, AcctClient: account.New(account.Params{DB: tx})})

	var accountIDs []string
	if accountID := ctx.Query("account_id"); accountID != "" {
		acct, err := helpers.GetOrgServiceAccount(requestCtx, orgID, accountID)
		if err != nil {
			return nil, err
		}
		accountIDs = []string{acct.ID}
	} else {
		accountIDs, err = helpers.OrgServiceAccountIDs(requestCtx, orgID, true, true)
		if err != nil {
			return nil, err
		}
	}

	tokens := []app.Token{}
	if len(accountIDs) > 0 {
		query := tx.
			Where(app.Token{
				OrgID:     orgID,
				TokenType: app.TokenTypeStatic,
			}).
			Where("account_id IN ?", accountIDs)
		if q := ctx.Query("q"); q != "" {
			query = query.Where("name ILIKE ? OR id = ?", "%"+q+"%", q)
		}

		res := query.
			Order("created_at DESC, id DESC").
			Scopes(scopes.WithOffsetPagination).
			Find(&tokens)
		if res.Error != nil {
			return nil, fmt.Errorf("unable to list tokens: %w", res.Error)
		}
	}

	tokens, err = db.HandlePaginatedResponse(ctx, tokens)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return enrichTokensWithAccounts(tx, orgID, tokens)
}

func enrichTokensWithAccounts(tx *gorm.DB, orgID string, tokens []app.Token) ([]TokenResponse, error) {
	resp := make([]TokenResponse, 0, len(tokens))
	if len(tokens) == 0 {
		return resp, nil
	}

	accountIDSet := make(map[string]struct{}, len(tokens))
	for _, token := range tokens {
		accountIDSet[token.AccountID] = struct{}{}
	}
	accountIDs := make([]string, 0, len(accountIDSet))
	for id := range accountIDSet {
		accountIDs = append(accountIDs, id)
	}

	var accounts []app.Account
	if err := tx.
		Where("id IN ?", accountIDs).
		Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("unable to load accounts: %w", err)
	}

	accountsByID := make(map[string]app.Account, len(accounts))
	for _, acct := range accounts {
		accountsByID[acct.ID] = acct
	}
	bindings, err := authz.New(authz.Params{DB: tx}).AccountRoleBindings(tx.Statement.Context, orgID, accountIDs)
	if err != nil {
		return nil, err
	}

	for _, token := range tokens {
		item := TokenResponse{Token: token, AccountRoleBindings: bindings[token.AccountID]}
		if limits, err := permissions.DecodeCredentialLimits(token.Authorization); err == nil && len(limits.Ceiling) == 1 {
			item.Scope = &limits.Ceiling[0].Scope
		}
		if acct, ok := accountsByID[token.AccountID]; ok {
			item.AccountName = acct.Name
			item.AccountEmail = acct.Email
			for _, binding := range item.AccountRoleBindings {
				if binding.Scope.Kind == permissions.KindOrg {
					item.AccountRole = string(binding.RoleType)
				}
			}
		}
		resp = append(resp, item)
	}

	return resp, nil
}

// @ID						DeleteToken
// @Summary				Delete an API token for a service account in the current org
// @Description.markdown	delete_token.md
// @Param					token_id	path	string	true	"token ID"
// @Tags					accounts
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				204
// @Router					/v1/tokens/{token_id} [DELETE]
func (s *service) DeleteToken(ctx *gin.Context) (require.Response, error) {
	access, err := require.MutationAccessFromContext(ctx)
	if err != nil {
		return require.Response{}, err
	}
	tx, target := access.DB(), access.Target()
	mutationCtx := tx.Statement.Context
	var token app.Token
	res := tx.Where(app.Token{
		ID: ctx.Param("token_id"), OrgID: target.OrgID, TokenType: app.TokenTypeStatic,
	}).First(&token)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return require.Response{}, stderr.ErrNotFound{Err: fmt.Errorf("token not found")}
	}
	if res.Error != nil {
		return require.Response{}, fmt.Errorf("unable to look up token: %w", res.Error)
	}
	authzClient := authz.New(authz.Params{DB: tx})
	acctClient := account.New(account.Params{DB: tx, AuthzClient: authzClient})
	helpers := accountshelpers.New(accountshelpers.Params{DB: tx, AuthzClient: authzClient, AcctClient: acctClient})
	if _, err := helpers.GetOrgServiceAccount(mutationCtx, target.OrgID, token.AccountID); err != nil {
		var notFound stderr.ErrNotFound
		var unauthorized stderr.ErrAuthorization
		if errors.As(err, &notFound) || errors.As(err, &unauthorized) {
			return require.Response{}, stderr.ErrNotFound{Err: fmt.Errorf("token not found")}
		}
		return require.Response{}, err
	}
	if err := acctClient.RevokeToken(mutationCtx, token.ID); err != nil {
		return require.Response{}, fmt.Errorf("unable to delete token: %w", err)
	}

	return require.Response{Status: http.StatusNoContent}, nil
}
