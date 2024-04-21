package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"gorm.io/gorm"
)

// @ID GetOrgInvites
// @Summary	Return org invites
// @Description.markdown get_org_invites.md
// @Tags			orgs
// @Accept			json
// @Produce		json
// @Security APIKey
// @Param   limit  query int	 false	"limit of health checks to return"	     Default(60)
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{array}		app.OrgInvite
// @Router			/v1/orgs/current/invites [GET]
func (s *service) GetOrgInvites(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	limitStr := ctx.DefaultQuery("limit", "60")
	limitVal, err := strconv.Atoi(limitStr)
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid limit %s: %w", limitStr, err),
			Description: "invalid limit",
		})
		return
	}

	orgs, err := s.getOrgInvites(ctx, org.ID, limitVal)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, orgs)
}

func (s *service) getOrgInvites(ctx context.Context, orgID string, limit int) ([]app.OrgInvite, error) {
	var org *app.Org

	res := s.db.WithContext(ctx).
		Preload("Invites", func(db *gorm.DB) *gorm.DB {
			return db.Order("org_invites.created_at DESC").Limit(limit)
		}).
		First(&org, "id = ?", orgID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get invites: %w", res.Error)
	}

	return org.Invites, nil
}
