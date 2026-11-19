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
	// GET /v1/roles?context=api_token. defaults to org_read_only. must be
	// empty for personal tokens.
	Role string `json:"role"`

	// "service_account" (default) creates a dedicated service account with
	// the given role; "personal" issues the token against your own account
	// and its existing roles, across all your orgs.
	TokenIdentity string `json:"token_identity"`
}

const (
	TokenIdentityPersonal       = "personal"
	TokenIdentityServiceAccount = "service_account"
)

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
// @Description			Deprecated: create a service account with POST /v1/service-accounts, then a token with POST /v1/tokens. Personal tokens (token_identity "personal") are deprecated because they carry all of the creator's access. Creates a long-lived static API token. By default (token_identity "service_account") each token gets its own dedicated service account and only grants access to the current org; the role param controls the token's permissions (any role assignable to API tokens; see GET /v1/roles?context=api_token) and defaults to org_read_only. With token_identity "personal" the token is issued against your own account instead: it uses your account's existing roles, is not limited to the current org, and the role param must be empty.
// @Param					req	body	CreateStaticTokenRequest	true	"Input"
// @Tags					accounts
// @Security				APIKey
// @Security				OrgID
// @Accept					json
// @Produce				json
// @Success				201	{object}	StaticTokenResponse
// @Deprecated				true
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

	org, err := s.helpers.RequireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	caller, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	token, err := s.issueStaticToken(ctx, org.ID, caller, &req, duration)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create static token: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, StaticTokenResponse{
		ID:       token.ID,
		APIToken: token.Token,
	})
}

func (s *service) issueStaticToken(ctx *gin.Context, orgID string, caller *app.Account, req *CreateStaticTokenRequest, duration time.Duration) (*app.Token, error) {
	switch req.TokenIdentity {
	case TokenIdentityPersonal:
		if req.Role != "" {
			return nil, stderr.NewInvalidRequest(fmt.Errorf("role cannot be set for personal tokens; the token uses your account's existing roles"))
		}
		return s.acctClient.IssueStaticToken(ctx, account.StaticTokenRequest{AccountID: caller.ID, OrgID: orgID, CreatedByID: caller.ID, Name: req.Name, Role: app.RoleTypeOrgAdmin, Duration: duration})
	case "", TokenIdentityServiceAccount:
		role, err := s.resolveTokenRole(ctx, orgID, req.Role)
		if err != nil {
			return nil, stderr.NewInvalidRequest(err)
		}
		return s.acctClient.CreateDedicatedStaticToken(ctx, account.StaticTokenRequest{OrgID: orgID, CreatedByID: caller.ID, Name: req.Name, Role: role, Duration: duration})
	default:
		return nil, stderr.NewInvalidRequest(fmt.Errorf("invalid token_identity %q: must be %q or %q", req.TokenIdentity, TokenIdentityPersonal, TokenIdentityServiceAccount))
	}
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
	org, err := s.helpers.RequireOrgAdmin(ctx)
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

	org, err := s.helpers.RequireOrgAdmin(ctx)
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

	if err := s.acctClient.DeleteStaticToken(ctx, org.ID, token.ID); err != nil {
		ctx.Error(fmt.Errorf("unable to delete static token: %w", err))
		return
	}

	ctx.Status(http.StatusNoContent)
}
