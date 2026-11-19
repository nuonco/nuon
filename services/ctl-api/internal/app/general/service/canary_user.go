package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
)

const (
	defaultCanaryAPITokenTimeout time.Duration = time.Minute
)

type CreateCanaryUserRequest struct {
	CanaryID string `json:"canary_id"`
}

type CreateCanaryUserResponse struct {
	APIToken        string `json:"api_token,omitzero"`
	GithubInstallID string `json:"github_install_id,omitzero"`
	Email           string `json:"email,omitzero"`
}

// @ID						CreateCanaryUser
// @Summary				create a temp user for running a canary
// @Description.markdown	create_canary_user.md
// @Param					req	body	CreateCanaryUserRequest	true	"Input"
// @Tags					general/admin
// @Accept					json
// @Produce				json
// @Success				201	{object}	CreateCanaryUserResponse
// @Deprecated
// @Router					/v1/general/canary-user [post]
func (s *service) CreateCanaryUser(ctx *gin.Context) {
	var req CreateCanaryUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	acct, token, err := s.createCanaryUser(ctx, req.CanaryID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create canary user: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, CreateCanaryUserResponse{
		APIToken:        token.Token,
		GithubInstallID: s.cfg.IntegrationGithubInstallID,
		Email:           acct.Email,
	})
}

func (s *service) createCanaryUser(ctx context.Context, canaryID string) (*app.Account, *app.Token, error) {
	acct, err := s.acctClient.FindAccount(ctx, canaryID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, err
		}
		acct, err = s.acctClient.CreateTestAccount(ctx, canaryID, app.AccountTypeCanary)
		if err != nil {
			return nil, nil, err
		}
	}

	token, err := s.acctClient.IssueToken(ctx, acct.ID, time.Hour, account.IssueTokenRequest{
		TokenType: app.TokenTypeCanary, Issuer: canaryID, CreatedByID: canaryID,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create canary user: %w", err)
	}

	return acct, token, nil
}
