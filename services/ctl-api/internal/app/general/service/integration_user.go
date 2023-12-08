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
	defaultIntegrationAPITokenTimeout time.Duration = time.Minute
)

type CreateIntegrationUserRequest struct{}

type CreateIntegrationUserResponse struct {
	APIToken        string `json:"api_token"`
	GithubInstallID string `json:"github_install_id"`
}

//	@BasePath	/v1/general
//
// create a temp user for running the integration test
//
//	@Summary	create a temp user for running integration test
//	@Schemes
//	@Description	create a temp user for running integration test
//	@Param			req	body	CreateIntegrationUserRequest	true	"Input"
//	@Tags			general/admin
//	@Accept			json
//	@Produce		json
//	@Success		201	{string}	ok
//	@Router			/v1/general/integration-user [post]
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

func (s *service) createIntegrationUser(ctx context.Context) (*app.UserToken, error) {
	intID := domains.NewIntegrationUserID()
	token := app.UserToken{
		CreatedByID: intID,
		Token:       intID,
		Subject:     intID,
		ExpiresAt:   time.Now().Add(defaultIntegrationAPITokenTimeout),
		IssuedAt:    time.Now(),
		Issuer:      intID,
		Email:       intID,
	}

	res := s.db.WithContext(ctx).
		Create(&token)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create integration user: %w", res.Error)
	}

	return &token, nil
}
