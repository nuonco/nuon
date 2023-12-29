package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
	"gorm.io/gorm/clause"
)

type CreateOrgUserRequest struct {
	UserID string `json:"user_id"`
}

// @ID AddUser
// @Summary	Add a user to the current org
// @Description.markdown	create_org_user.md
// @Param			req	body	CreateOrgUserRequest	true	"Input"
// @Tags			orgs
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		201				{object}	app.UserOrg
// @Router			/v1/orgs/current/user [POST]
func (s *service) CreateUser(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req CreateOrgUserRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	userOrg, err := s.createUser(ctx, org.ID, req.UserID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create user: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, userOrg)
}

func (s *service) createUser(ctx context.Context, orgID, userID string) (*app.UserOrg, error) {
	userOrg := &app.UserOrg{
		OrgID:  orgID,
		UserID: userID,
	}

	err := s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&userOrg).Error
	if err != nil {
		return nil, fmt.Errorf("unable to add user to org: %w", err)
	}
	return userOrg, nil
}
