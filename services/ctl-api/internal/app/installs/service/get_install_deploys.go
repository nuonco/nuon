package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetInstallDeploys
// @Summary				get all deploys to an install
// @Description.markdown	get_install_deploys.md
// @Param					install_id					path	string	true	"install ID"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
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
// @Success				200	{array}		app.InstallDeploy
// @Router					/v1/installs/{install_id}/deploys [GET]
func (s *service) GetInstallDeploys(ctx *gin.Context) {
	appID := ctx.Param("install_id")

	installDeploys, err := s.getInstallDeploys(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install deploys: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installDeploys)
}

func (s *service) getInstallDeploys(ctx *gin.Context, installID string) ([]*app.InstallDeploy, error) {
	var installDeploys []*app.InstallDeploy
	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Preload("ActionWorkflowRuns").
		Preload("InstallComponent").
		Preload("InstallComponent.Component").
		Preload("ComponentBuild").
		Preload("ComponentBuild.ComponentConfigConnection").
		Preload("ComponentBuild.VCSConnectionCommit").
		Joins("JOIN install_components ON install_components.id=install_deploys.install_component_id").
		Where("install_components.install_id = ?", installID).
		Order("created_at desc").
		Find(&installDeploys)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install deploys: %w", res.Error)
	}

	installDeploys, err := db.HandlePaginatedResponse(ctx, installDeploys)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return installDeploys, nil
}
