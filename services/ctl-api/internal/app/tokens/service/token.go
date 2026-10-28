package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type CreateTokenRequest struct {
	AccountID string `json:"account_id" validate:"required"`

	// Name labels the token where it is listed; defaults to the account's identity.
	Name string `json:"name"`

	// Duration defaults to one year.
	Duration string `json:"duration" default:"8760h"`

	Invalidate bool `json:"invalidate"`
}

type CreateTokenResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	AccountID string    `json:"account_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	Token     string    `json:"token"`
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
func (s *service) CreateToken(ctx *gin.Context) {
	org, err := s.helpers.RequireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req CreateTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse request: %w", err),
			Description: fmt.Sprintf("unable to parse request: %s", err.Error()),
		})
		return
	}

	if req.AccountID == "" {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("account_id is required"),
			Description: "account_id is required",
		})
		return
	}

	token, err := s.helpers.CreateServiceAccountToken(ctx, org, req.AccountID, req.Name, req.Duration, req.Invalidate)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, CreateTokenResponse{
		ID:        token.ID,
		Name:      token.Name,
		AccountID: token.AccountID,
		ExpiresAt: token.ExpiresAt,
		CreatedAt: token.CreatedAt,
		Token:     token.Token,
	})
}

// TokenResponse is app.Token enriched with the owning account's identity and
// its current role in this org (the stored Token.Role is a stale copy from
// creation time).
type TokenResponse struct {
	app.Token

	AccountName  string `json:"account_name"`
	AccountEmail string `json:"account_email"`
	AccountRole  string `json:"account_role"`
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
func (s *service) ListTokens(ctx *gin.Context) {
	org, err := s.helpers.RequireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var accountIDs []string
	if accountID := ctx.Query("account_id"); accountID != "" {
		acct, err := s.helpers.GetOrgServiceAccount(ctx, org.ID, accountID)
		if err != nil {
			ctx.Error(err)
			return
		}
		accountIDs = []string{acct.ID}
	} else {
		accountIDs, err = s.helpers.OrgServiceAccountIDs(ctx, org.ID, true, true)
		if err != nil {
			ctx.Error(err)
			return
		}
	}

	tokens := []app.Token{}
	if len(accountIDs) > 0 {
		tx := s.db.WithContext(ctx).
			Where(app.Token{
				OrgID:     org.ID,
				TokenType: app.TokenTypeStatic,
			}).
			Where("account_id IN ?", accountIDs)
		if q := ctx.Query("q"); q != "" {
			tx = tx.Where("name ILIKE ? OR id = ?", "%"+q+"%", q)
		}

		res := tx.
			Order("created_at DESC").
			Scopes(scopes.WithOffsetPagination).
			Find(&tokens)
		if res.Error != nil {
			ctx.Error(fmt.Errorf("unable to list tokens: %w", res.Error))
			return
		}
	}

	tokens, err = db.HandlePaginatedResponse(ctx, tokens)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to handle paginated response: %w", err))
		return
	}

	resp, err := s.enrichTokensWithAccounts(ctx, org.ID, tokens)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to load token accounts: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// enrichTokensWithAccounts loads the owning accounts for a page of tokens in
// a single query (preloading only the Roles for the current org) and merges
// each account's name, email, and current org role onto its token.
func (s *service) enrichTokensWithAccounts(ctx *gin.Context, orgID string, tokens []app.Token) ([]TokenResponse, error) {
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
	if err := s.db.WithContext(ctx).
		Preload("Roles", "org_id = ?", orgID).
		Preload("Roles.Org").
		Where("id IN ?", accountIDs).
		Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("unable to load accounts: %w", err)
	}

	accountsByID := make(map[string]app.Account, len(accounts))
	for _, acct := range accounts {
		accountsByID[acct.ID] = acct
	}

	for _, token := range tokens {
		item := TokenResponse{Token: token}
		if acct, ok := accountsByID[token.AccountID]; ok {
			item.AccountName = acct.Name
			item.AccountEmail = acct.Email
			item.AccountRole = string(accountshelpers.OrgRoleType(&acct, orgID))
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
func (s *service) DeleteToken(ctx *gin.Context) {
	org, err := s.helpers.RequireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var token app.Token
	res := s.db.WithContext(ctx).
		Where(app.Token{
			ID:        ctx.Param("token_id"),
			OrgID:     org.ID,
			TokenType: app.TokenTypeStatic,
		}).
		First(&token)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		ctx.Error(stderr.ErrNotFound{Err: fmt.Errorf("token not found")})
		return
	}
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to look up token: %w", res.Error))
		return
	}

	if _, err := s.helpers.GetOrgServiceAccount(ctx, org.ID, token.AccountID); err != nil {
		var notFound stderr.ErrNotFound
		var unauthorized stderr.ErrAuthorization
		if errors.As(err, &notFound) || errors.As(err, &unauthorized) {
			ctx.Error(stderr.ErrNotFound{Err: fmt.Errorf("token not found")})
			return
		}
		ctx.Error(err)
		return
	}

	if err := s.acctClient.RevokeToken(ctx, token.ID); err != nil {
		ctx.Error(fmt.Errorf("unable to delete token: %w", err))
		return
	}

	ctx.Status(http.StatusNoContent)
}
