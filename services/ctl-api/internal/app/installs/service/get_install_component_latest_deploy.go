package service

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetInstallComponentLatestDeploy
// @Summary				get the latest deploy for an install component
// @Description.markdown	get_install_component_latest_deploy.md
// @Param					install_id		path	string	true	"install ID"
// @Param					component_id	path	string	true	"component ID"
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
// @Router					/v1/installs/{install_id}/components/{component_id}/deploys/latest [get]
func (s *service) GetInstallComponentLatestDeploy(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	componentID := ctx.Param("component_id")

	installDeploy, err := s.getInstallComponentLatestDeploy(ctx, installID, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install component latest deploy: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installDeploy)
}

func (s *service) getInstallComponentLatestDeploy(ctx context.Context, installID, componentID string) (app.InstallDeploy, error) {
	ginCtx, ok := ctx.(*gin.Context)
	if !ok {
		return app.InstallDeploy{}, fmt.Errorf("invalid context type, expected *gin.Context")
	}
	deploys, err := s.getInstallComponentDeploys(ginCtx, installID, componentID)
	if err != nil {
		return app.InstallDeploy{}, fmt.Errorf("unable to get install component deploys: %w", err)
	}

	filterOutPlanOnly := make([]app.InstallDeploy, 0, len(deploys))
	for _, deploy := range deploys {
		if deploy.PlanOnly {
			continue
		}
		filterOutPlanOnly = append(filterOutPlanOnly, deploy)
	}

	if len(filterOutPlanOnly) == 0 {
		return app.InstallDeploy{}, nil
	}

	sort.Slice(filterOutPlanOnly, func(i, j int) bool {
		return filterOutPlanOnly[i].CreatedAt.After(filterOutPlanOnly[j].CreatedAt)
	})

	return filterOutPlanOnly[0], nil
}
