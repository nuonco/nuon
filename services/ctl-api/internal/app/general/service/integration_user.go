package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
)

const (
	defaultIntegrationAPITokenTimeout time.Duration = time.Minute * 10
)

type CreateIntegrationUserRequest struct{}

type CreateIntegrationUserResponse struct {
	APIToken        string `json:"api_token,omitzero"`
	GithubInstallID string `json:"github_install_id,omitzero"`
	Email           string `json:"email,omitzero"`
}

// @ID						CreateIntegrationUser
// @Summary				create a temp user for running integration test
// @Description.markdown	create_integration_user.md
// @Param					req	body	CreateIntegrationUserRequest	true	"Input"
// @Tags					general/admin
// @Accept					json
// @Produce				json
// @Success				201	{object}	CreateIntegrationUserResponse
// @Router					/v1/general/integration-user [post]
func (s *service) CreateIntegrationUser(ctx *gin.Context) {
	token, email, err := s.createIntegrationUser(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create integration user: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, CreateIntegrationUserResponse{
		APIToken:        token.Token,
		GithubInstallID: s.cfg.IntegrationGithubInstallID,
		Email:           email,
	})
}

func (s *service) createIntegrationUser(ctx context.Context) (*app.Token, string, error) {
	intID := domains.NewIntegrationUserID()
	acct, err := s.acctClient.CreateTestAccount(ctx, intID, app.AccountTypeIntegration)
	if err != nil {
		return nil, "", err
	}

	token, err := s.acctClient.IssueToken(ctx, acct.ID, defaultIntegrationAPITokenTimeout, account.IssueTokenRequest{
		TokenType: app.TokenTypeIntegration, Issuer: intID, CreatedByID: intID,
	})
	if err != nil {
		return nil, "", fmt.Errorf("unable to create integration user: %w", err)
	}

	return token, acct.Email, nil
}
