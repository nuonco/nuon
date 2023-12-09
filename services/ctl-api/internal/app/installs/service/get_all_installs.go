package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

//	@BasePath	/v1/installs
//
// Get all installs
//
//	@Summary	get all installs for all orgs
//	@Schemes
//	@Description	get all installs
//	@Tags			installs/admin
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}	app.Install
//	@Router			/v1/installs [get]
func (s *service) GetAllInstalls(ctx *gin.Context) {
	installs, err := s.getAllInstalls(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get installs for: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, installs)
}

func (s *service) getAllInstalls(ctx context.Context) ([]*app.Install, error) {
	var installs []*app.Install
	res := s.db.WithContext(ctx).
		Preload("AppSandboxConfig").
		Preload("AWSAccount").
		Preload("App").
		Preload("App.AppSandboxConfigs").
		Order("created_at desc").
		Find(&installs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get all installs: %w", res.Error)
	}

	return installs, nil
}
