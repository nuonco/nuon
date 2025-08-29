package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
	"gorm.io/gorm"
)

// @ID						GetOrgInvites
// @Summary				Return org invites
// @Description.markdown	get_org_invites.md
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					orgs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.OrgInvite
// @Router					/v1/orgs/current/invites [GET]
func (s *service) GetOrgInvites(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	// TODO: remove when pagination is enabled
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

func (s *service) getOrgInvites(ctx *gin.Context, orgID string, limit int) ([]app.OrgInvite, error) {
	var org *app.Org

	res := s.db.WithContext(ctx).
		Preload("Invites", func(db *gorm.DB) *gorm.DB {
			return db.
				Scopes(scopes.WithOffsetPagination).
				Order("org_invites.created_at DESC").Limit(limit)
		}).
		First(&org, "id = ?", orgID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get invites: %w", res.Error)
	}

	invites, err := db.HandlePaginatedResponse(ctx, org.Invites)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	org.Invites = invites
	return org.Invites, nil
}
