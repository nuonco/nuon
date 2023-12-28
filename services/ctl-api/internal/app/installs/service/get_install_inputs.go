package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

// @ID GetInstallInputs
// @Summary	get an installs inputs
// @Description.markdown	get_install_inputs.md
// @Tags			installs
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Param			install_id		path		string	true	"install ID"
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{array}		app.InstallInputs
// @Router			/v1/installs/{install_id}/inputs [GET]
func (s *service) GetInstallInputs(ctx *gin.Context) {
	appID := ctx.Param("install_id")

	installInputs, err := s.getInstallInputs(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install inputs: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installInputs)
}

func (s *service) getInstallInputs(ctx context.Context, installID string) ([]app.InstallInputs, error) {
	var install app.Install
	res := s.db.WithContext(ctx).
		Preload("InstallInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_inputs.created_at DESC")
		}).
		First(&install, "id = ?", installID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install inputs: %w", res.Error)
	}

	return install.InstallInputs, nil
}
