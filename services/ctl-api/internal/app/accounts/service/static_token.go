package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateStaticTokenRequest struct {
	Duration      string `json:"duration" default:"8760h"`
	Name          string `json:"name" validate:"required"`
	Role          string `json:"role"`
	TokenIdentity string `json:"token_identity"`
}

const (
	TokenIdentityPersonal       = "personal"
	TokenIdentityServiceAccount = "service_account"
)

// @ID						CreateStaticToken
// @Summary				create a static API token for your org
// @Description			Retired: returns 410 Gone. Create a service account with POST /v1/service-accounts, then issue a token with an explicit role and scope using POST /v1/tokens. Existing tokens can still be listed and revoked.
// @Param					req	body	CreateStaticTokenRequest	true	"Input"
// @Tags					accounts
// @Security				APIKey
// @Security				OrgID
// @Accept					json
// @Produce				json
// @Failure				410	{object}	stderr.ErrResponse
// @Deprecated				true
// @Router					/v1/account/static-token [POST]
func (s *service) CreateStaticToken(ctx *gin.Context) {
	ctx.JSON(http.StatusGone, gin.H{"error": "token creation retired; use POST /v1/tokens with account_id, role and scope"})
}

// @ID						ListStaticTokens
// @Summary				list your org's static API tokens
// @Description			Deprecated: use GET /v1/tokens instead. Lists the static API tokens for your current org. Token secrets are never returned.
// @Tags					accounts
// @Security				APIKey
// @Security				OrgID
// @Produce				json
// @Success				200	{array}	app.Token
// @Deprecated				true
// @Router					/v1/account/static-tokens [GET]
func (s *service) ListStaticTokens(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	if _, err := require.CustomerRead(ctx, s.db, cctx.GetLogger(ctx, s.l), authz.ReadRequest{
		Permission: permissions.OrgWrite,
		Target:     permissions.Scope{Kind: permissions.KindOrg, OrgID: org.ID, ID: org.ID},
	}); err != nil {
		ctx.Error(err)
		return
	}

	var tokens []app.Token
	res := s.db.WithContext(ctx).
		Where(app.Token{
			OrgID:     org.ID,
			TokenType: app.TokenTypeStatic,
		}).
		Order("created_at DESC").
		Find(&tokens)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to list static tokens: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}

// @ID						DeleteStaticToken
// @Summary				delete a static API token
// @Description			Deprecated: use DELETE /v1/tokens/{token_id} instead. Deletes a static API token belonging to your current org. For service account tokens, the dedicated service account is deleted as well; for personal tokens, only the token is deleted and your account is untouched. Once deleted, the token can no longer be used to access the API.
// @Param					token_id	path	string	true	"token ID"
// @Tags					accounts
// @Security				APIKey
// @Security				OrgID
// @Produce				json
// @Success				204
// @Deprecated				true
// @Router					/v1/account/static-tokens/{token_id} [DELETE]
func (s *service) DeleteStaticToken(ctx *gin.Context) {
	tokenID := ctx.Param("token_id")

	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	err = require.CustomerMutation(ctx, s.db, cctx.GetLogger(ctx, s.l), authz.MutationRequest{
		Permission: permissions.OrgWrite,
		Target:     permissions.Scope{Kind: permissions.KindOrg, OrgID: org.ID, ID: org.ID},
	}, func(tx *gorm.DB, target permissions.ResolvedTarget) error {
		mutationCtx := tx.Statement.Context
		var token app.Token
		res := tx.Where(app.Token{ID: tokenID, OrgID: target.OrgID, TokenType: app.TokenTypeStatic}).First(&token)
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return stderr.ErrNotFound{Err: fmt.Errorf("token not found")}
		}
		if res.Error != nil {
			return fmt.Errorf("unable to look up static token: %w", res.Error)
		}
		authzClient := authz.New(authz.Params{DB: tx})
		acctClient := account.New(account.Params{DB: tx, AuthzClient: authzClient})
		if err := acctClient.RevokeToken(mutationCtx, token.ID); err != nil {
			return fmt.Errorf("unable to delete static token: %w", err)
		}
		scoped := &service{db: tx, authzClient: authzClient, acctClient: acctClient}
		return scoped.deleteTokenServiceAccount(mutationCtx, target.OrgID, token.AccountID)
	})
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// dedicatedTokenSubjectPrefix marks an account that exists only to hold one static
// token, so deleting the token deletes the account.
func dedicatedTokenSubjectPrefix(orgID string) string {
	return fmt.Sprintf("%s-token-", orgID)
}

func (s *service) deleteTokenServiceAccount(ctx context.Context, orgID, accountID string) error {
	var acct app.Account
	res := s.db.WithContext(ctx).Where(app.Account{ID: accountID}).First(&acct)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil
	}
	if res.Error != nil {
		return fmt.Errorf("unable to look up token account: %w", res.Error)
	}

	// never remove roles from or delete a personal token's real user account
	if acct.AccountType != app.AccountTypeService {
		return nil
	}

	// Only accounts created to hold this one token: revoking a stack's or runner's token
	// must not take the identity with it.
	if !strings.HasPrefix(acct.Subject, dedicatedTokenSubjectPrefix(orgID)) {
		return nil
	}

	if err := authz.RequireUserManagedInOrg(s.db.WithContext(ctx), accountID, orgID); err != nil {
		return err
	}
	return s.acctClient.DeleteServiceAccountByID(ctx, accountID)
}
