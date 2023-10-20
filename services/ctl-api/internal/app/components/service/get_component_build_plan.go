package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/pkg/generics"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

//	@BasePath	/v1/components
//
// Get install build plan
//
//	@Summary	get component build plan
//	@Schemes
//	@Description	get component build plan
//	@Param			component_id	path	string	true	"component ID"
//	@Param			build_id	path	string	true	"build ID"
//	@Tags components
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
//	@Router			/v1/components/{component_id}/builds/{build_id}/plan [get]
func (s *service) GetComponentBuildPlan(ctx *gin.Context) {
	componentID := ctx.Param("component_id")
	buildID := ctx.Param("build_id")

	plan, err := s.getComponentBuildPlan(ctx, componentID, buildID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install build plan: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, plan)
}

func (s *service) getComponentBuildPlan(ctx context.Context, componentID, buildID string) (*planv1.Plan, error) {
	plan := generics.GetFakeObj[*planv1.WaypointPlan]()
	return &planv1.Plan{
		Actual: &planv1.Plan_WaypointPlan{
			WaypointPlan: plan,
		},
	}, nil
}
