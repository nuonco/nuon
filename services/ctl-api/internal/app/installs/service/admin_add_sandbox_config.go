package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type AdminAddSandboxConfigInstallRequest struct{}

// Deprovision an install
//
//	@Summary	deprovision an install, but keep it in the database
//
//	@Schemes
//
//	@Description	deprovision an install
//
//
//	@Tags			installs/admin
//	@Accept			json
//	@Param			req			body	AdminAddSandboxConfigInstallRequest	true	"Input"
//
//
//	@Produce		json
//	@Success		201	{string}	ok
//	@Router			/v1/installs/admin-add-sandbox-configs [POST]
func (s *service) AdminAddSandboxConfigInstall(ctx *gin.Context) {
	installs, err := s.getAllInstalls(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get all installs: %w", err))
		return
	}

	for _, install := range installs {
		if install.AppSandboxConfigID != "" {
			continue
		}

		if len(install.App.AppSandboxConfigs) < 1 {
			ctx.Error(fmt.Errorf("app does not have sandbox configs: %s", install.AppID))
			return
		}

		res := s.db.WithContext(ctx).
			Model(&install).
			Updates(app.Install{AppSandboxConfigID: install.App.AppSandboxConfigs[0].ID})
		if res.Error != nil {
			ctx.Error(fmt.Errorf("unable to update install: %s", install.AppID))
			return
		}
	}

	ctx.JSON(http.StatusOK, true)
}
