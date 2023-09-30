package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

//	@BasePath	/v1/releases
//
// Get all releases
//
//	@Summary	get all releases for all orgs
//	@Schemes
//	@Description	get all installs
//	@Tags			releases/admin
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}	app.ComponentRelease
//	@Router			/v1/releases [get]
func (s *service) GetAllReleases(ctx *gin.Context) {
	installs, err := s.getAllInstalls(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get installs for: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, installs)
}

func (s *service) getAllInstalls(ctx context.Context) ([]*app.Install, error) {
	var installs []*app.Install
	res := s.db.WithContext(ctx).Find(&installs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get all installs: %w", res.Error)
	}

	return installs, nil
}
