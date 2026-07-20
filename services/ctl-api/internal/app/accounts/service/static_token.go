package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateStaticTokenRequest struct {
	// defaults to one year
	Duration string `json:"duration" default:"8760h"`

	// human-friendly name to identify the token later
	Name string `json:"name" validate:"required"`
}

type StaticTokenResponse struct {
	ID       string `json:"id,omitzero"`
	APIToken string `json:"api_token,omitzero"`
}

// @ID						CreateStaticToken
// @Summary				create a static API token for your org's service account
// @Description			Creates a long-lived static API token scoped to your current org. The token is issued for the org's service account, which is created automatically if it does not already exist. The token only grants access to the current org.
// @Param					req	body	CreateStaticTokenRequest	true	"Input"
// @Tags					accounts
// @Security				APIKey
// @Security				OrgID
// @Accept					json
// @Produce				json
// @Success				201	{object}	StaticTokenResponse
// @Router					/v1/account/static-token [POST]
func (s *service) CreateStaticToken(ctx *gin.Context) {
	var req CreateStaticTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	if req.Name == "" {
		ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("name is required")))
		return
	}

	if req.Duration == "" {
		req.Duration = "8760h"
	}
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("invalid duration: %w", err)))
		return
	}
	if duration <= 0 {
		ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("duration must be positive")))
		return
	}

	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	caller, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	acct, err := s.ensureServiceAccount(ctx, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to ensure service account: %w", err))
		return
	}

	token, err := s.createStaticToken(ctx, acct, caller.ID, req.Name, duration)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create static token: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, StaticTokenResponse{
		ID:       token.ID,
		APIToken: token.Token,
	})
}

// @ID						ListStaticTokens
// @Summary				list your org's static API tokens
// @Description			Lists the static API tokens for your current org's service account. Token secrets are never returned.
// @Tags					accounts
// @Security				APIKey
// @Security				OrgID
// @Produce				json
// @Success				200	{array}	app.Token
// @Router					/v1/account/static-tokens [GET]
func (s *service) ListStaticTokens(ctx *gin.Context) {
	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	acct, err := s.ensureServiceAccount(ctx, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to ensure service account: %w", err))
		return
	}

	var tokens []app.Token
	res := s.db.WithContext(ctx).
		Where(app.Token{
			AccountID: acct.ID,
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
// @Description			Deletes a static API token belonging to your current org's service account. Once deleted, the token can no longer be used to access the API.
// @Param					token_id	path	string	true	"token ID"
// @Tags					accounts
// @Security				APIKey
// @Security				OrgID
// @Produce				json
// @Success				204
// @Router					/v1/account/static-tokens/{token_id} [DELETE]
func (s *service) DeleteStaticToken(ctx *gin.Context) {
	tokenID := ctx.Param("token_id")

	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	acct, err := s.ensureServiceAccount(ctx, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to ensure service account: %w", err))
		return
	}

	res := s.db.WithContext(ctx).
		Where(app.Token{
			ID:        tokenID,
			AccountID: acct.ID,
			TokenType: app.TokenTypeStatic,
		}).
		Delete(&app.Token{})
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to delete static token: %w", res.Error))
		return
	}
	if res.RowsAffected == 0 {
		ctx.Error(stderr.ErrNotFound{Err: fmt.Errorf("token not found")})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (s *service) requireOrgAdmin(ctx *gin.Context) (*app.Org, error) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		return nil, err
	}

	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		return nil, err
	}

	if !s.isOrgAdmin(acct, org.ID) {
		return nil, stderr.ErrAuthorization{
			Err:         fmt.Errorf("only org admins can manage static tokens"),
			Description: "only org admins can manage static tokens",
		}
	}

	return org, nil
}

func (s *service) isOrgAdmin(acct *app.Account, orgID string) bool {
	for _, role := range acct.Roles {
		if role.RoleType == app.RoleTypeOrgAdmin && role.OrgID.ValueString() == orgID {
			return true
		}
	}
	return false
}

func (s *service) ensureServiceAccount(ctx context.Context, orgID string) (*app.Account, error) {
	name := fmt.Sprintf("%s-admin-service-account", orgID)
	email := account.ServiceAccountEmail(name)

	acct, err := s.acctClient.FindAccount(ctx, email)
	if err == nil {
		return acct, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("unable to lookup account: %w", err)
	}

	newAcct := app.Account{
		Email:       email,
		Subject:     name,
		AccountType: app.AccountTypeService,
	}
	if res := s.db.WithContext(ctx).Create(&newAcct); res.Error != nil {
		return nil, fmt.Errorf("unable to create service account: %w", res.Error)
	}

	if err := s.authzClient.AddAccountOrgRole(ctx, app.RoleTypeOrgAdmin, orgID, newAcct.ID); err != nil {
		return nil, fmt.Errorf("unable to add org role to service account: %w", err)
	}

	return &newAcct, nil
}

func (s *service) createStaticToken(ctx context.Context, acct *app.Account, createdByID, name string, duration time.Duration) (*app.Token, error) {
	token := app.Token{
		CreatedByID: createdByID,
		Name:        name,
		Token:       domains.NewUserTokenID(),
		TokenType:   app.TokenTypeStatic,
		ExpiresAt:   time.Now().Add(duration),
		IssuedAt:    time.Now(),
		Issuer:      acct.ID,
		AccountID:   acct.ID,
	}

	if res := s.db.WithContext(ctx).Create(&token); res.Error != nil {
		return nil, fmt.Errorf("unable to create static token: %w", res.Error)
	}

	return &token, nil
}
