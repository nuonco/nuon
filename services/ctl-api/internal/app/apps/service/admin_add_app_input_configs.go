package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type AdminAddAppInputsConfigs struct{}

//	@BasePath	/v1/apps
//
//	 Add app inputs for all apps
//
//	@Summary	add inputs for all apps
//	@Schemes
//	@Description	add app inputs for all apps
//	@Param			req	body	AdminAddAppInputsConfigs	true	"Input"
//	@Tags			apps/admin
//	@Accept			json
//	@Produce		json
//	@Success		200	{boolean}	true
//	@Router			/v1/apps/admin-add-app-inputs [POST]
func (s *service) AdminAddAppInputsConfigs(ctx *gin.Context) {
	var req AdminAddAppInputsConfigs
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	apps, err := s.getAllApps(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to fetch apps: %w", err))
		return
	}

	for _, app := range apps {
		if err := s.adminCreateAppInputsConfigs(ctx, app); err != nil {
			ctx.Error(fmt.Errorf("unable to create app inputs: %w", err))
			return
		}
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) adminCreateAppInputsConfigs(ctx context.Context, currentApp *app.App) error {
	appInputsConfigs := app.AppInputConfig{
		CreatedByID: currentApp.CreatedByID,
		OrgID:       currentApp.OrgID,
		AppID:       currentApp.ID,
		AppInputs:   []app.AppInput{},
	}

	res := s.db.WithContext(ctx).Create(&appInputsConfigs)
	if res.Error != nil {
		return fmt.Errorf("unable to create app inputs %w", res.Error)
	}

	return nil
}
