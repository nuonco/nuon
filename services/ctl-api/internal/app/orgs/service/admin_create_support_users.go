package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

var defaultSupportUsers = []string{
	// Dre Smith
	"google-oauth2|113884954942864770921",
	//Jon Morehouse
	"google-oauth2|114670241124324496631",
	//Jordan Acosta
	"google-oauth2|106750268626571499868",
	//Nat Hamilton
	"google-oauth2|107796233904597398271",
}

// @ID AdminCreateSupportUsers
// @BasePath	/v1/orgs
// @Summary	Add nuon users as support members
// @Description.markdown create_org_support_users.md
// @Param			org_id	path	string	true	"org ID for your current org"
// @Tags			orgs/admin
// @Accept			json
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/orgs/{org_id}/admin-support-users [POST]
func (s *service) CreateSupportUsers(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	cctx := context.WithValue(ctx, "user_id", org.CreatedByID)

	if err := s.ensureUsers(ctx, defaultSupportUsers); err != nil {
		ctx.Error(fmt.Errorf("unable to ensure users: %w", err))
		return
	}

	// add each user to this org
	for _, userID := range defaultSupportUsers {
		if _, err := s.createUser(cctx, orgID, userID); err != nil {
			ctx.Error(fmt.Errorf("unable to add users to org: %w", err))
			return
		}
	}

	ctx.JSON(http.StatusCreated, map[string]string{
		"status": "ok",
	})
}

func (s *service) ensureUsers(ctx context.Context, userIDs []string) error {
	// make sure each user exists in the database first
	for _, userID := range defaultSupportUsers {
		_, err := s.getUser(ctx, userID)
		if err == nil {
			continue
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		token := app.UserToken{
			CreatedByID: userID,
			Token:       domains.NewUserTokenID(),
			Subject:     userID,
			ExpiresAt:   time.Now(),
			IssuedAt:    time.Now(),
			Issuer:      userID,
			Email:       userID,
			TokenType:   app.TokenTypeAdmin,
		}

		res := s.db.WithContext(ctx).
			Create(&token)
		if res.Error != nil {
			return fmt.Errorf("unable to create user: %w", res.Error)
		}
	}

	return nil
}

func (s *service) getUser(ctx context.Context, subject string) (*app.UserToken, error) {
	var user app.UserToken

	res := s.db.WithContext(ctx).Where(app.UserToken{
		Subject: subject,
	}).First(&user)
	if res.Error != nil {
		return nil, res.Error
	}

	return &user, nil
}
