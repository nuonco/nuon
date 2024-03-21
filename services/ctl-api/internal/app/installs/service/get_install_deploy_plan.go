package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/powertoolsdev/mono/pkg/workflows/dal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
)

// @ID GetInstallDeployPlan
// @Summary	get install deploy plan
// @Description.markdown	get_install_deploy_plan.md
// @Param			install_id	path	string	true	"install ID"
// @Param			deploy_id	path	string	true	"deploy ID"
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
// @Success		200				{object} planv1.Plan
// @Router			/v1/installs/{install_id}/deploys/{deploy_id}/plan [get]
func (s *service) GetInstallDeployPlan(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")
	deployID := ctx.Param("deploy_id")

	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install: %w", err))
		return
	}

	deploy, err := s.getInstallDeploy(ctx, installID, deployID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install deploy: %w %s", err, deployID))
		return
	}

	plan, err := s.getInstallDeployPlan(ctx,
		org.ID,
		install.AppID,
		deploy.ComponentBuild.ComponentConfigConnection.ComponentID,
		deployID,
		installID, deploy.Type)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install deploy plan: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, plan)
}

func (s *service) getInstallDeployPlan(ctx context.Context, orgID, appID, componentID, deployID, installID string, deployTyp app.InstallDeployType) (*planv1.Plan, error) {
	wkflowDal, err := dal.New(s.v, dal.WithSettings(dal.Settings{
		DeploymentsBucket:                s.orgsOutputs.Buckets.Deployments.Name,
		DeploymentsBucketIAMRoleTemplate: s.orgsOutputs.OrgsIAMRoleNameTemplateOutputs.DeploymentsAccess,
	}), dal.WithOrgID(orgID))
	if err != nil {
		return nil, fmt.Errorf("unable to get dal for deploy plan: %w", err)
	}

	var (
		plan *planv1.Plan
	)
	switch deployTyp {
	case app.InstallDeployTypeTeardown:
		plan, err = wkflowDal.GetInstanceDestroyPlan(ctx, orgID, appID, componentID, deployID, installID)
	default:
		plan, err = wkflowDal.GetInstanceDeployPlan(ctx, orgID, appID, componentID, deployID, installID)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get plan: %w", err)
	}

	return plan, nil
}
