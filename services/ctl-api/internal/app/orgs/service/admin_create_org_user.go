package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm/clause"
)

type AdminCreateOrgUserRequest struct {
	Email string `json:"email"`
}

// @ID AdminAddOrgUser
// @BasePath	/v1/orgs
// @Summary	Add a user to an org
// @Description.markdown create_org_user.md
// @Param			org_id	path	string	true	"org ID to add user too"
// @Tags			orgs/admin
// @Param			req	body	AdminCreateOrgUserRequest	true	"Input"
// @Accept			json
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/orgs/{org_id}/admin-add-user [POST]
func (s *service) CreateOrgUser(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org: %w", err))
		return
	}

	var req AdminCreateOrgUserRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	userOrg, err := s.createUserByEmail(ctx, org, req.Email)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create user: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, userOrg)
}

func (s *service) createUserByEmail(ctx context.Context, org *app.Org, email string) (*app.UserOrg, error) {
	var userToken app.UserToken
	err := s.db.WithContext(ctx).Order("created_at DESC").First(&userToken, "email = ?", email).Error
	if err != nil {
		return nil, fmt.Errorf("unable to add user to org: %w", err)
	}

	userOrg := &app.UserOrg{
		CreatedByID: org.CreatedByID,
		OrgID:       org.ID,
		UserID:      userToken.Subject,
	}

	err = s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&userOrg).Error
	if err != nil {
		return nil, fmt.Errorf("unable to add user to org: %w", err)
	}
	return userOrg, nil
}
