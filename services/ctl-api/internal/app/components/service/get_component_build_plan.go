package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/powertoolsdev/mono/pkg/workflows/dal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
)

// @ID GetComponentBuildPlan
// @Summary	get component build plan
// @Description.markdown	get_component_build_plan.md
// @Param			component_id	path	string	true	"component ID"
// @Param			build_id		path	string	true	"build ID"
// @Tags			components
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{object}	planv1.Plan
// @Router			/v1/components/{component_id}/builds/{build_id}/plan [get]
func (s *service) GetComponentBuildPlan(ctx *gin.Context) {
	org, err := middlewares.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	componentID := ctx.Param("component_id")
	component, err := s.getComponent(ctx, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component: %w", err))
		return
	}

	buildID := ctx.Param("build_id")

	plan, err := s.getComponentBuildPlan(ctx, org.ID, component.AppID, componentID, buildID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install build plan: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, plan)
}

func (s *service) getComponentBuildPlan(ctx context.Context, orgID, appID, componentID, buildID string) (*planv1.Plan, error) {
	wkflowDal, err := dal.New(s.v, dal.WithSettings(dal.Settings{
		DeploymentsBucket:                s.orgsOutputs.Buckets.Deployments.Name,
		DeploymentsBucketIAMRoleTemplate: s.orgsOutputs.OrgsIAMRoleNameTemplateOutputs.DeploymentsAccess,
	}), dal.WithOrgID(orgID))
	if err != nil {
		return nil, fmt.Errorf("unable to get build plan: %w", err)
	}

	plan, err := wkflowDal.GetBuildPlan(ctx, orgID, appID, componentID, buildID)
	if err != nil {
		return nil, fmt.Errorf("unable to get build plan: %w", err)
	}

	return plan, nil
}
