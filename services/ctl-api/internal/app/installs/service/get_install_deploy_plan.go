package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

//	@ID						GetInstallDeployPlan
//	@Summary				get install deploy plan
//	@Description.markdown	get_install_deploy_plan.md
//	@Param					install_id	path	string	true	"install ID"
//	@Param					deploy_id	path	string	true	"deploy ID"
//	@Tags					installs
//	@Accept					json
//	@Produce				json
//	@Security				APIKey
//	@Security				OrgID
//	@Failure				400	{object}	stderr.ErrResponse
//	@Failure				401	{object}	stderr.ErrResponse
//	@Failure				403	{object}	stderr.ErrResponse
//	@Failure				404	{object}	stderr.ErrResponse
//	@Failure				500	{object}	stderr.ErrResponse
//	@Success				200	{object}	planv1.Plan
//	@Router					/v1/installs/{install_id}/deploys/{deploy_id}/plan [get]
func (s *service) GetInstallDeployPlan(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	deployID := ctx.Param("deploy_id")

	deploy, err := s.getInstallDeploy(ctx, installID, deployID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install deploy: %w %s", err, deployID))
		return
	}

	if len(deploy.RunnerJobs) < 1 {
		ctx.Error(stderr.ErrNotReady{
			Err:         errors.New("runner job is not ready yet"),
			Description: "runner job is not ready yet",
		})
		return
	}

	plan, err := s.getRunnerJobPlan(ctx, deploy.RunnerJobs[0].ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install deploy plan: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, plan)
}

func (s *service) getRunnerJobPlan(ctx context.Context, runnerJobID string) (*planv1.Plan, error) {
	var runnerJobPlan app.RunnerJobPlan
	res := s.db.WithContext(ctx).
		First(&runnerJobPlan, "runner_job_id = ?", runnerJobID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get runner job plan")
	}

	plan, err := apiPlanToProto([]byte(runnerJobPlan.PlanJSON))
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse plan")
	}

	return plan, nil
}

func apiPlanToProto(byts []byte) (*planv1.Plan, error) {
	plan := &planv1.Plan{}
	if err := protojson.Unmarshal(byts, plan); err != nil {
		return nil, fmt.Errorf("unable to unmarshal plan bytes into proto: %w", err)
	}

	return plan, nil
}
