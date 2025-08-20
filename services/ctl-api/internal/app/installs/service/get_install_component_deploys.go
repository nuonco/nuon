package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
	"gorm.io/gorm"
)

// @ID						GetInstallComponentDeploys
// @Summary				get an install components deploys
// @Description.markdown	get_install_component_deploys.md
// @Param					install_id					path	string	true	"install ID"
// @Param					component_id				path	string	true	"component ID"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Param					x-nuon-pagination-enabled	header	bool	false	"Enable pagination"
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
	install := app.InstallComponent{
		ComponentID: componentID,
		InstallID:   installID,
	}
	res := s.db.WithContext(ctx).Preload("InstallDeploys", func(db *gorm.DB) *gorm.DB {
		return db.
			Scopes(scopes.WithOffsetPagination).
			Order("install_deploys.created_at DESC").Limit(1000)
	}).
		Preload("InstallDeploys.CreatedBy").
		Preload("InstallDeploys.ComponentBuild").
		Preload("InstallDeploys.ComponentBuild.VCSConnectionCommit").
		Preload("TerraformWorkspace").
		Where(install).
		First(&install)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install component: %w", res.Error)
	}

	workflowIDs := make([]string, 0, len(install.InstallDeploys))
	for _, deploy := range install.InstallDeploys {
		workflowID := deploy.WorkflowID
		if workflowID != nil {
			workflowIDs = append(workflowIDs, *workflowID)
		}
	}

	planStatusMap, err := s.helpers.GetWorkflowsPlanOnlyMap(ctx, workflowIDs)
	if err != nil {
		return nil, fmt.Errorf("unable to get workflow plan status: %w", err)
	}

	filterOutPlanOnly := make([]app.InstallDeploy, 0, len(install.InstallDeploys))
	for i := range install.InstallDeploys {
		workflowID := install.InstallDeploys[i].WorkflowID
		if planOnly, exists := planStatusMap[*workflowID]; exists {
			if !planOnly {
				filterOutPlanOnly = append(filterOutPlanOnly, install.InstallDeploys[i])
			}
		}
	}

	dpls, err := db.HandlePaginatedResponse(ctx, filterOutPlanOnly)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	install.InstallDeploys = dpls

	return install.InstallDeploys, nil
}
