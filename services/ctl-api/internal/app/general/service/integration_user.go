package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

const (
	defaultIntegrationAPITokenTimeout time.Duration = time.Minute * 10
)

type CreateIntegrationUserRequest struct{}

type CreateIntegrationUserResponse struct {
	APIToken        string `json:"api_token"`
	GithubInstallID string `json:"github_install_id"`
	Email           string `json:"email"`
}

// @ID CreateIntegrationUser
// @Summary	create a temp user for running integration test
// @Description.markdown create_integration_user.md
// @Param			req	body	CreateIntegrationUserRequest	true	"Input"
// @Tags			general/admin
// @Accept			json
// @Produce		json
// @Success		201	{object} CreateIntegrationUserResponse
// @Router			/v1/general/integration-user [post]
func (s *service) CreateIntegrationUser(ctx *gin.Context) {
	token, err := s.createIntegrationUser(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create integration user: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, CreateIntegrationUserResponse{
		APIToken:        token.Token,
		GithubInstallID: s.cfg.IntegrationGithubInstallID,
	})
}

func (s *service) createIntegrationUser(ctx context.Context) (*app.Token, error) {
	intID := domains.NewIntegrationUserID()

	acct := app.Account{
		Email:       fmt.Sprintf("%s@nuon.co", intID),
		Subject:     intID,
		AccountType: app.AccountTypeIntegration,
	}
	res := s.db.WithContext(ctx).
		Create(&acct)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create integration account: %w", res.Error)
	}

	token := app.Token{
		CreatedByID: intID,
		AccountID:   acct.ID,
		Token:       domains.NewUserTokenID(),
		TokenType:   app.TokenTypeIntegration,
		ExpiresAt:   time.Now().Add(defaultIntegrationAPITokenTimeout),
		IssuedAt:    time.Now(),
		Issuer:      intID,
	}

	res = s.db.WithContext(ctx).
		Create(&token)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create integration user: %w", res.Error)
	}

	return &token, nil
}
