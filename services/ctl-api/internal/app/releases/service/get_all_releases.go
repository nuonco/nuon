package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID GetAllReleases
// @Summary	get all releases for all orgs
// @Description.markdown	get all releases for all orgs
// @Tags			releases/admin
// @Accept			json
// @Produce		json
// @Success		200	{array}	app.ComponentRelease
// @Router			/v1/releases [get]
func (s *service) GetAllReleases(ctx *gin.Context) {
	releases, err := s.getAllReleases(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get all releases: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, releases)
}

func (s *service) getAllReleases(ctx context.Context) ([]*app.ComponentRelease, error) {
	var releases []*app.ComponentRelease
	res := s.db.WithContext(ctx).Find(&releases)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get all releases: %w", res.Error)
	}

	return releases, nil
}
