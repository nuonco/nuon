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
	defaultCanaryAPITokenTimeout time.Duration = time.Minute
)

type CreateCanaryUserRequest struct {
	CanaryID string `json:"canary_id"`
}

type CreateCanaryUserResponse struct {
	APIToken        string `json:"api_token"`
	GithubInstallID string `json:"github_install_id"`
}

// @ID CreateCanaryUser
// @Summary	create a temp user for running a canary
// @Description.markdown create_canary_user.md
// @Param			req	body	CreateCanaryUserRequest	true	"Input"
// @Tags			general/admin
// @Accept			json
// @Produce		json
// @Success		201	{object} CreateCanaryUserResponse
// @Router			/v1/general/canary-user [post]
func (s *service) CreateCanaryUser(ctx *gin.Context) {
	var req CreateCanaryUserRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request input: %w", err))
		return
	}

	token, err := s.createCanaryUser(ctx, req.CanaryID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create integration user: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, CreateCanaryUserResponse{
		APIToken:        token.Token,
		GithubInstallID: s.cfg.IntegrationGithubInstallID,
	})
}

func (s *service) createCanaryUser(ctx context.Context, canaryID string) (*app.UserToken, error) {
	token := app.UserToken{
		CreatedByID: canaryID,
		Token:       domains.NewUserTokenID(),
		Subject:     canaryID,
		ExpiresAt:   time.Now().Add(time.Hour),
		IssuedAt:    time.Now(),
		Issuer:      canaryID,
		Email:       canaryID,
	}

	res := s.db.WithContext(ctx).
		Create(&token)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create integration user: %w", res.Error)
	}

	return &token, nil
}
