package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetAllReleases
// @Summary				get all releases for all orgs
// @Description.markdown	get all releases for all orgs
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					releases/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{array}	app.ComponentRelease
// @Router					/v1/releases [get]
func (s *service) GetAllReleases(ctx *gin.Context) {
	releases, err := s.getAllReleases(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get all releases: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, releases)
}

func (s *service) getAllReleases(ctx *gin.Context) ([]*app.ComponentRelease, error) {
	var releases []*app.ComponentRelease
	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Find(&releases)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get all releases: %w", res.Error)
	}

	releases, err := db.HandlePaginatedResponse(ctx, releases)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return releases, nil
}
