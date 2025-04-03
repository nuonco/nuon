package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

// @ID						DeleteInstallComponent
// @Summary				delete an install component
// @Description.markdown	delete_install_component.md
// @Param					install_id		path	string				true	"install ID"
// @Param					component_id	path	string				true	"component ID"
// @Param					force					query	bool					false	"force delete"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{boolean} 		true
// @Router					/v1/installs/{install_id}/components/{component_id} [delete]
func (s *service) DeleteInstallComponent(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	componentID := ctx.Param("component_id")
	force := ctx.DefaultQuery("force", "false") == "true"

	deploy, err := s.queueDeleteInstallDeploy(ctx, installID, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to delete install component: %w", err))
		return
	}

	s.evClient.Send(ctx, installID, &signals.Signal{
		Type:        signals.OperationDeploy,
		DeployID:    deploy.ID,
		ForceDelete: force,
	})
	ctx.JSON(http.StatusOK, true)
}

func (s *service) queueDeleteInstallDeploy(ctx *gin.Context, installID, componentID string) (*app.InstallDeploy, error) {
	var installDeploy app.InstallDeploy
	res := s.db.WithContext(ctx).
		Joins("JOIN install_components ON install_components.id = install_deploys.install_component_id").
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
		Status:             app.InstallDeployStatusQueued,
		StatusDescription:  "waiting to be deployed to install",
		ComponentBuildID:   installDeploy.ComponentBuildID,
		InstallComponentID: installDeploy.InstallComponentID,
		Type:               app.InstallDeployTypeTeardown,
	}
	res = s.db.WithContext(ctx).
		Create(&teardownDeploy)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to queue install deploy: %w", res.Error)
	}

	res = s.db.WithContext(ctx).
		Model(&app.InstallComponent{}).
		Where("component_id = ? and install_id = ?", componentID, installID).
		Update("status", app.InstallDeployStatusQueued)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to update install component: %w", res.Error)
	}

	return &teardownDeploy, nil
}
