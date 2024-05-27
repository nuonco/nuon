package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

// TODO: add this back as a param to the spec.
// Either make this a DELETE method (which it arguably should be,)
// or get a fix merged to the elixir openapi SDK generate to handle empty request types.
type TeardownInstallComponentRequest struct{}

func (c *TeardownInstallComponentRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID TeardownInstallComponent
// @Summary	teardown an install component
// @Description.markdown	teardown_install_component.md
// @Param			install_id	path	string						true	"install ID"
// @Param			component_id	path	string	true	"component ID"
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
// @Success		201				{object}	app.InstallDeploy
// @Router			/v1/installs/{install_id}/components/{component_id}/teardown [post]
func (s *service) TeardownInstallComponent(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	componentID := ctx.Param("component_id")

	var req TeardownInstallComponentRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	deploy, err := s.teardownInstallDeploy(ctx, installID, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to teardown install component: %w", err))
		return
	}

	s.evClient.Send(ctx, installID, &signals.Signal{
		Type:     signals.OperationDeploy,
		DeployID: deploy.ID,
	})
	ctx.JSON(http.StatusCreated, deploy)
}

func (s *service) teardownInstallDeploy(ctx context.Context, installID, componentID string) (*app.InstallDeploy, error) {
	var installDeploy app.InstallDeploy
	res := s.db.WithContext(ctx).
		Joins("JOIN install_components ON install_components.id = install_deploys.install_component_id").
		Preload("InstallComponent").
		Preload("ComponentBuild").
		Order("created_at desc").
		Where("install_components.component_id = ?", componentID).
		Where("install_components.install_id = ?", installID).
		Limit(1).
		First(&installDeploy)
	if res.Error != nil {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("unable to get previous install deploy: %w", res.Error),
			Description: "please make sure this install component has been deployed, before tearing down",
		}
	}

	teardownDeploy := app.InstallDeploy{
		Status:             "queued",
		StatusDescription:  "waiting to be deployed to install",
		ComponentBuildID:   installDeploy.ComponentBuildID,
		InstallComponentID: installDeploy.InstallComponentID,
		Type:               app.InstallDeployTypeTeardown,
	}
	res = s.db.WithContext(ctx).
		Create(&teardownDeploy)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to add teardown install deploy: %w", res.Error)
	}

	return &teardownDeploy, nil
}
