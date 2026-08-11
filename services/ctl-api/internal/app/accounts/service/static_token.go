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

	// org role granted to the token. must be assignable to API tokens; see
	// GET /v1/roles?context=api_token. defaults to org_read_only.
	Role string `json:"role"`
}

const defaultTokenRole = app.RoleTypeOrgReadOnly

func (s *service) resolveTokenRole(ctx *gin.Context, orgID, raw string) (app.RoleType, error) {
	if raw == "" {
		raw = string(defaultTokenRole)
	}
	resolved, err := s.authzClient.ResolveAssignableRole(ctx, orgID, app.RoleType(raw), app.RoleContextAPIToken)
	if err != nil {
		return "", err
	}
	return resolved.RoleType, nil
}

type StaticTokenResponse struct {
	ID       string `json:"id,omitzero"`
	APIToken string `json:"api_token,omitzero"`
}

const defaultTokenDuration = "8760h"

func parseTokenDuration(raw string) (time.Duration, error) {
	if raw == "" {
		raw = defaultTokenDuration
	}
	duration, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid duration: %w", err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("duration must be positive")
	}
	return duration, nil
}

// @ID						CreateStaticToken
// @Summary				create a static API token for your org
// @Description			Creates a long-lived static API token scoped to your current org. Each token gets its own dedicated service account, and only grants access to the current org. The role param controls the token's permissions (any role assignable to API tokens; see GET /v1/roles?context=api_token) and defaults to org_read_only.
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

	duration, err := parseTokenDuration(req.Duration)
	if err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	role, err := s.resolveTokenRole(ctx, org.ID, req.Role)
	if err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	caller, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	acct, err := s.createTokenServiceAccount(ctx, org.ID, role)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create service account: %w", err))
		return
	}

	token, err := s.createStaticToken(ctx, acct, org.ID, caller.ID, req.Name, role, duration)
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
// @Description			Lists the static API tokens for your current org. Token secrets are never returned.
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
// @Description			Deletes a static API token belonging to your current org, along with its dedicated service account. Once deleted, the token can no longer be used to access the API.
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

	var token app.Token
	res := s.db.WithContext(ctx).
		Where(app.Token{
			ID:        tokenID,
			OrgID:     org.ID,
			TokenType: app.TokenTypeStatic,
		}).
		First(&token)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		ctx.Error(stderr.ErrNotFound{Err: fmt.Errorf("token not found")})
		return
	}
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to look up static token: %w", res.Error))
		return
	}

	if err := s.db.WithContext(ctx).Delete(&token).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to delete static token: %w", err))
		return
	}

	if err := s.deleteTokenServiceAccount(ctx, org.ID, token.AccountID); err != nil {
		ctx.Error(fmt.Errorf("unable to delete service account: %w", err))
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

func (s *service) createTokenServiceAccount(ctx context.Context, orgID string, roleType app.RoleType) (*app.Account, error) {
	name := fmt.Sprintf("%s-token-%s", orgID, domains.NewAccountID())
	email := account.ServiceAccountEmail(name)

	newAcct := app.Account{
		Email:       email,
		Subject:     name,
		AccountType: app.AccountTypeService,
	}
	if res := s.db.WithContext(ctx).Create(&newAcct); res.Error != nil {
		return nil, fmt.Errorf("unable to create service account: %w", res.Error)
	}

	if err := s.authzClient.AddAccountOrgRole(ctx, roleType, orgID, newAcct.ID); err != nil {
		return nil, fmt.Errorf("unable to add org role to service account: %w", err)
	}

	return &newAcct, nil
}

func (s *service) deleteTokenServiceAccount(ctx context.Context, orgID, accountID string) error {
	if err := s.authzClient.RemoveAccountOrgRoles(ctx, orgID, accountID); err != nil {
		return fmt.Errorf("unable to remove service account roles: %w", err)
	}

	if err := s.db.WithContext(ctx).
		Where(app.Account{ID: accountID, AccountType: app.AccountTypeService}).
		Delete(&app.Account{}).Error; err != nil {
		return fmt.Errorf("unable to delete service account: %w", err)
	}

	return nil
}

func (s *service) createStaticToken(ctx context.Context, acct *app.Account, orgID, createdByID, name string, role app.RoleType, duration time.Duration) (*app.Token, error) {
	token := app.Token{
		CreatedByID: createdByID,
		Name:        name,
		Role:        string(role),
		OrgID:       orgID,
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
