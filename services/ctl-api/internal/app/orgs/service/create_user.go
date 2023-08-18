package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm/clause"
)

type CreateUserRequest struct {
	UserID string `json:"user_id"`
}

// @BasePath /v1/orgs/

// Add a user to an org
// @Summary Add a user to an org
// @Schemes
// @Description add a user to an org
// @Param org_id path string true "org ID for your current org"
// @Param req body CreateUserRequest true "Input"
// @Tags orgs
// @Accept json
// @Produce json
// @Success 201 {object} app.UserOrg
// @Router /v1/orgs/{org_id}/user [POST]
func (s *service) CreateUser(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req CreateUserRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	if err := s.createUser(ctx, orgID, req.UserID); err != nil {
		ctx.Error(fmt.Errorf("unable to create user: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, map[string]string{
		"status": "ok",
	})
}

func (s *service) createUser(ctx context.Context, orgID, userID string) error {
	userOrg := &app.UserOrg{
		OrgID:  orgID,
		UserID: userID,
	}

	err := s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&userOrg).Error
	if err != nil {
		return fmt.Errorf("unable to add user to org: %w", err)
	}
	return nil
}
