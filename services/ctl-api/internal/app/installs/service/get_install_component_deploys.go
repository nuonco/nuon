package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetInstallComponentDeploys
// @Summary				get an install components deploys
// @Description.markdown	get_install_component_deploys.md
// @Param					install_id					path	string	true	"install ID"
// @Param					component_id				path	string	true	"component ID"
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
// @Router					/v1/installs/{install_id}/components/{component_id}/deploys [GET]
func (s *service) GetInstallComponentDeploys(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	componentID := ctx.Param("component_id")

	installComponentDeploys, err := s.getInstallComponentDeploys(ctx, installID, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install component deploys: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installComponentDeploys)
}

func (s *service) getInstallComponentDeploys(ctx *gin.Context, installID, componentID string) ([]app.InstallDeploy, error) {
	installComp, err := s.getInstallComponent(ctx, installID, componentID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install component")
	}

	var installDeploys []app.InstallDeploy
	res := s.db.WithContext(ctx).
		Where(app.InstallDeploy{
			InstallComponentID: installComp.ID,
		}).
		Scopes(scopes.WithOffsetPagination).
		Preload("CreatedBy").
		Preload("ComponentBuild").
		Preload("ComponentBuild.VCSConnectionCommit").
		Joins("JOIN install_workflows ON install_workflows.id=install_deploys.install_workflow_id").
		Where("install_workflows.plan_only = ?", false).
		Order("created_at DESC").
		Limit(100).
		First(&installDeploys)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install deploys: %w", res.Error)
	}

	dpls, err := db.HandlePaginatedResponse(ctx, installDeploys)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return dpls, nil
}
