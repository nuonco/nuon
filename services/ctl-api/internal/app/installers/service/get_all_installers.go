package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID GetAllInstallers
// @Summary	get all installers for all orgs
// @Description.markdown	get_all_installers.md
// @Tags installers/admin
// @Accept			json
// @Produce		json
// @Success		200	{array}	app.AppInstaller
// @Router			/v1/installers [get]
func (s *service) GetAllAppInstallers(ctx *gin.Context) {
	apps, err := s.getAllAppInstallers(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get apps for: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, apps)
}

func (s *service) getAllAppInstallers(ctx context.Context) ([]*app.AppInstaller, error) {
	var apps []*app.AppInstaller
	res := s.db.WithContext(ctx).
		Preload("Metadata").
		Find(&apps)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get all app installers: %w", res.Error)
	}
	return apps, nil
}
