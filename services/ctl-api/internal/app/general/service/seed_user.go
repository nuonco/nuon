package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
)

const (
	defaultSeedAPITokenTimeout time.Duration = time.Hour * 24 * 30
	defaultSeedUserName        string        = "seed"
)

type CreateSeedUserRequest struct{}

type CreateSeedUserResponse struct {
	APIToken        string `json:"api_token,omitzero"`
	GithubInstallID string `json:"github_install_id,omitzero"`
	Email           string `json:"email,omitzero"`
}

// @ID						CreateSeedUser
// @Summary				create a temp user for running integration test
// @Description.markdown	create_integration_user.md
// @Param					req	body	CreateSeedUserRequest	true	"Input"
// @Tags					general/admin
// @Accept					json
// @Produce				json
// @Success				201	{object}	CreateSeedUserResponse
// @Router					/v1/general/integration-user [post]
func (s *service) CreateSeedUser(ctx *gin.Context) {
	token, err := s.createSeedUser(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create integration user: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, CreateSeedUserResponse{
		APIToken:        token.Token,
		GithubInstallID: s.cfg.IntegrationGithubInstallID,
	})
}

func (s *service) createSeedUser(ctx context.Context) (*app.Token, error) {
	email := fmt.Sprintf("%s@nuon.co", defaultSeedUserName)
	acct, err := s.acctClient.FindAccount(ctx, email)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		acct, err = s.acctClient.CreateAccount(ctx, email, email, account.NoUserJourneys())
		if err != nil {
			return nil, err
		}
	}

	token := app.Token{
		CreatedByID: acct.ID,
		AccountID:   acct.ID,
		Token:       domains.NewUserTokenID(),
		TokenType:   app.TokenTypeAuth0,
		ExpiresAt:   time.Now().Add(defaultSeedAPITokenTimeout),
		IssuedAt:    time.Now(),
		Issuer:      "seed",
	}

	res := s.db.WithContext(ctx).
		Create(&token)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create seed user: %w", res.Error)
	}

	return &token, nil
}
