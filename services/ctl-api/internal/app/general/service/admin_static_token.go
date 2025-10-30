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
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type StaticTokenRequest struct {
	// defaults to one year
	Duration string `json:"duration" validate:"required" default:"8760h"`

	EmailOrSubject string `json:"email_or_subject" validate:"required"`
}

func (c *StaticTokenRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

type StaticTokenResponse struct {
	APIToken string `json:"api_token,omitzero"`
}

// @ID						AdminCreateStaticToken
// @Summary				create a static token for a user.
// @Description.markdown	create_static_token.md
// @Param					req	body	StaticTokenRequest	true	"Input"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				201	{object}	StaticTokenResponse
// @Router					/v1/general/admin-static-token [POST]
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

	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		ctx.Error(fmt.Errorf("invalid time duration: %w", err))
		return
	}

	token, err := s.createStaticToken(ctx, req.EmailOrSubject, duration)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create static token: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, StaticTokenResponse{
		APIToken: token.Token,
	})
}

func (s *service) createStaticToken(ctx context.Context, subjectOrEmail string, duration time.Duration) (*app.Token, error) {
	acct, err := s.acctClient.FindAccount(ctx, subjectOrEmail)
	if err != nil {
		return nil, fmt.Errorf("invalid account: %w", err)
	}

	token := app.Token{
		CreatedByID: acct.ID,
		Token:       domains.NewUserTokenID(),
		TokenType:   app.TokenTypeStatic,
		ExpiresAt:   time.Now().Add(duration),
		IssuedAt:    time.Now(),
		Issuer:      acct.ID,
		AccountID:   acct.ID,
	}

	res := s.db.WithContext(ctx).
		Create(&token)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create static token: %w", res.Error)
	}

	return &token, nil
}
