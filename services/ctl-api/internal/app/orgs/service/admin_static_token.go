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
	"gorm.io/gorm/clause"
)

type StaticTokenRequest struct {
	// defaults to one year
	Duration string `json:"duration" validate:"required" default:"8760h"`
}

func (c *StaticTokenRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

type StaticTokenResponse struct {
	APIToken string `json:"api_token"`
}

// @ID AdminCreateStaticToken
// @Summary	create a static token with access to the org.
// @Description.markdown create_static_token.md
// @Param			req	body	StaticTokenRequest	true	"Input"
// @Param			org_id	path	string	true	"org ID for your current org"
// @Tags orgs/admin
// @Accept			json
// @Produce		json
// @Success		201	{object} StaticTokenResponse
// @Router			/v1/orgs/{org_id}/admin-static-token [POST]
func (s *service) AdminCreateStaticToken(ctx *gin.Context) {
	var req StaticTokenRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request input: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	orgID := ctx.Param("org_id")

	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		ctx.Error(fmt.Errorf("invalid time duration: %w", err))
		return
	}

	token, err := s.createStaticToken(ctx, orgID, duration)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create integration user: %w", err))
		return
	}

	_, err = s.createUser(ctx, orgID, token.Subject)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to add static token user to org: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, StaticTokenResponse{
		APIToken: token.Token,
	})
}

func (s *service) createStaticToken(ctx context.Context, orgID string, duration time.Duration) (*app.UserToken, error) {
	email := fmt.Sprintf("%s-static@nuon.co", orgID)
	token := app.UserToken{
		CreatedByID: email,
		Token:       domains.NewUserTokenID(),
		Subject:     email,
		ExpiresAt:   time.Now().Add(duration),
		IssuedAt:    time.Now(),
		Issuer:      email,
		Email:       email,
		TokenType:   app.TokenTypeStatic,
	}

	res := s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "subject"}},
			UpdateAll: true,
		}).
		Create(&token)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create static user token: %w", res.Error)
	}

	return &token, nil
}
