package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetInstallDeploy
// @Summary				get an install deploy
// @Description.markdown	get_install_deploy.md
// @Param					install_id	path	string	true	"install ID"
// @Param					deploy_id	path	string	true	"deploy ID"
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
// @Success				200	{object}	app.InstallDeploy
// @Router					/v1/installs/{install_id}/deploys/{deploy_id} [get]
func (s *service) GetInstallDeploy(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	deployID := ctx.Param("deploy_id")

	installDeploy, err := s.getInstallDeploy(ctx, installID, deployID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install deploy %s: %w", deployID, err))
		return
	}

	ctx.JSON(http.StatusOK, installDeploy)
}

func (s *service) getInstallDeploy(ctx context.Context, installID, deployID string) (*app.InstallDeploy, error) {
	var installDeploy app.InstallDeploy
	res := s.db.WithContext(ctx).
		Joins("JOIN install_components ON install_components.id=install_deploys.install_component_id").
		Preload("InstallComponent").
		Preload("InstallComponent.Component").
		Preload("RunnerJobs", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(scopes.WithDisableViews).Order("created_at desc").Limit(10)
		}).
		Preload("ActionWorkflowRuns").
		Preload("LogStream").
		Preload("OCIArtifact").
		Preload("ComponentBuild").
		Preload("ComponentBuild.ComponentConfigConnection").
		Preload("ComponentBuild.ComponentConfigConnection.Component").
    Preload("ComponentBuild.VCSConnectionCommit").
		Where("install_components.install_id = ?", installID).
		First(&installDeploy, "install_deploys.id = ?", deployID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install deploy: %w", res.Error)
	}

	return &installDeploy, nil
}
