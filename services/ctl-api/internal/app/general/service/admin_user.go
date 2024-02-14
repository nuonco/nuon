package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type AdminUserRequest struct {
	Email    string        `json:"email" validate:"required"`
	Duration time.Duration `json:"duration" validate:"required" default:"24h"`
}

func (c *AdminUserRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

type AdminUserResponse struct {
	APIToken string `json:"api_token"`
}

// @ID AdminUser
// @Summary	create an admin user for internal purposes, such as testing.
// @Description.markdown create_admin_user.md
// @Param			req	body	AdminUserRequest	true	"Input"
// @Tags			general/admin
// @Accept			json
// @Produce		json
// @Success		201	{object} AdminUserResponse
// @Router			/v1/general/admin-user [post]
func (s *service) CreateAdminUser(ctx *gin.Context) {
	var req AdminUserRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request input: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	token, err := s.createAdminUser(ctx, req.Email, req.Duration)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create integration user: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, AdminUserResponse{
		APIToken: token.Token,
	})
}

func (s *service) createAdminUser(ctx context.Context, email string, duration time.Duration) (*app.UserToken, error) {
	token := app.UserToken{
		CreatedByID: email,
		Token:       domains.NewUserTokenID(),
		Subject:     email,
		ExpiresAt:   time.Now().Add(duration),
		IssuedAt:    time.Now(),
		Issuer:      email,
		Email:       email,
		TokenType:   app.TokenTypeAdmin,
	}

	res := s.db.WithContext(ctx).
		Create(&token)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create integration user: %w", res.Error)
	}

	return &token, nil
}
