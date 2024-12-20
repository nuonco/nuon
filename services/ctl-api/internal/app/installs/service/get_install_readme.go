package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/pkg/render"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type Readme struct {
	ReadeMe string `json:"readme"`
}

// @ID GetInstallReadme
// @Summary	get install readme rendered with
// @Description.markdown	get_install_readme.md
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
// @Success		200				{object} Readme
// @Router			/v1/installs/{install_id}/readme [get]
func (s *service) GetInstallReadme(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	// 1. grab the install
	installID := ctx.Param("install_id")
	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install: %w", err))
		return
	}

	// 2. make sure we have one in hand
	appConfig, err := s.getLatestAppConfig(ctx, install.AppID)
	if err != nil {
		response := Readme{""}
		ctx.JSON(http.StatusOK, response)
		return
	}

	// 2. grab the latest successful deploy plan
	deploy, err := s.getInstallLatestSuccessfulDeploy(ctx, installID)
	if err != nil {
		response := Readme{appConfig.Readme}
		ctx.JSON(http.StatusOK, response)
		return
	}

	// 3. grab the plan
	plan, err := s.getInstallDeployPlan(ctx,
		org.ID,
		install.AppID,
		deploy.ComponentBuild.ComponentConfigConnection.ComponentID,
		deploy.ID,
		installID, deploy.Type)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install deploy plan: %w", err))
		return
	}

	// 4. interpolate the variables into the readme md
	vars := plan.GetWaypointPlan().Variables
	readme := appConfig.Readme
	value, err := render.RenderString(ctx, readme, vars.IntermediateData)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get render readme: %w", err))
		return
	}
	fmt.Printf(value)
	response := Readme{value}

	ctx.JSON(http.StatusOK, response)
}
func (s *service) getLatestAppConfig(ctx context.Context, appID string) (*app.AppConfig, error) {
	var appConfig app.AppConfig
	res := s.db.WithContext(ctx).Where("app_id = ?", appID).Order("created_at DESC").First(&appConfig)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app config: %w", res.Error)
	}
	return &appConfig, nil
}

func (s *service) getInstallLatestSuccessfulDeploy(ctx context.Context, installID string) (*app.InstallDeploy, error) {
	var installDeploy app.InstallDeploy
	res := s.db.WithContext(ctx).
		Joins("JOIN install_components ON install_components.id=install_deploys.install_component_id").
		Preload("ComponentBuild").
		Preload("ComponentBuild.ComponentConfigConnection").
		Preload("ComponentBuild.ComponentConfigConnection.Component").
		Where("install_components.install_id = ?", installID).
		First(&installDeploy, "install_deploys.status = ?", app.InstallDeployStatusOK)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install deploy: %w", res.Error)
	}

	return &installDeploy, nil
}
