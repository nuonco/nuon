package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/pkg/generics"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

//	@BasePath	/v1/installs
//
// Get install deploy plan
//
//	@Summary	get install deploy plan
//	@Schemes
//	@Description	get install deploy plan
//	@Param			install_id	path	string	true	"install ID"
//	@Param			deploy_id	path	string	true	"deploy ID"
//	@Tags			installs
//	@Accept			json
//	@Produce		json
//	@Param			X-Nuon-Org-ID	header		string	true	"org ID"
//	@Param			Authorization	header		string	true	"bearer auth token"
//	@Failure		400				{object}	stderr.ErrResponse
//	@Failure		401				{object}	stderr.ErrResponse
//	@Failure		403				{object}	stderr.ErrResponse
//	@Failure		404				{object}	stderr.ErrResponse
//	@Failure		500				{object}	stderr.ErrResponse
//	@Success		200				{object} planv1.Plan
//	@Router			/v1/installs/{install_id}/deploys/{deploy_id}/plan [get]
func (s *service) GetInstallDeployPlan(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	deployID := ctx.Param("deploy_id")

	plan, err := s.getInstallDeployPlan(ctx, installID, deployID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install deploy plan: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, plan)
}

func (s *service) getInstallDeployPlan(ctx context.Context, installID, componentID string) (*planv1.Plan, error) {
	plan := generics.GetFakeObj[*planv1.WaypointPlan]()
	return &planv1.Plan{
		Actual: &planv1.Plan_WaypointPlan{
			WaypointPlan: plan,
		},
	}, nil
}
