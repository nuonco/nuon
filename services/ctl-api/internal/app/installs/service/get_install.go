package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
)

// @ID GetInstall
// @Summary	get an install
// @Description.markdown	get_install.md
// @Param			install_id	path	string	true	"install ID"
// @Tags			installs
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{object}	app.Install
// @Router			/v1/installs/{install_id} [get]
func (s *service) GetInstall(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")

	install, err := s.findInstall(ctx, org.ID, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installID, err))
		return
	}

	ctx.JSON(http.StatusOK, install)
}

func (s *service) findInstall(ctx context.Context, orgID, installID string) (*app.Install, error) {
	install := app.Install{}
	res := s.db.WithContext(ctx).
		Preload("AWSAccount").
		Preload("App").
		Preload("App.Org").
		Preload("AppSandboxConfig").
		Preload("AppRunnerConfig").
		Where("name = ? AND org_id = ?", installID, orgID).
		Or("id = ?", installID).
		First(&install)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	return &install, nil
}
