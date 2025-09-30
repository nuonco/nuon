package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetRunnerJobCompositePlan
// @Summary				get runner job composite plan
// @Description.markdown	get_runner_job_composite_plan.md
// @Param					runner_job_id	path	string	true	"runner job ID"
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	plantypes.CompositePlan
// @Router					/v1/runner-jobs/{runner_job_id}/composite-plan [get]
func (s *service) GetRunnerJobCompositePlan(ctx *gin.Context) {
	runnerJobID := ctx.Param("runner_job_id")

	cp, err := s.getRunnerJobCompositePlan(ctx, runnerJobID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner job: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, cp)
}

func (s *service) getRunnerJobCompositePlan(ctx context.Context, runnerJobID string) (*plantypes.CompositePlan, error) {
	var runnerPlan app.RunnerJobPlan

	res := s.db.WithContext(ctx).
		Where(app.RunnerJobPlan{
			RunnerJobID: runnerJobID,
		}).First(&runnerPlan)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get job plan: %w", res.Error)
	}

	var runnerJob app.RunnerJob
	res = s.db.WithContext(ctx).
		Where(app.RunnerJob{
			ID: runnerJobID,
		}).First(&runnerJob)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get job: %w", res.Error)
	}

	var compositePlan plantypes.CompositePlan
	switch runnerJob.Group {
	case app.RunnerJobGroupSync:
		err := json.Unmarshal([]byte(runnerPlan.PlanJSON), &compositePlan.SyncPlan)
		if err != nil {
			return nil, fmt.Errorf("unable to unmarshal sync oci plan: %w", err)
		}
	case app.RunnerJobGroupBuild:
		err := json.Unmarshal([]byte(runnerPlan.PlanJSON), &compositePlan.BuildPlan)
		if err != nil {
			return nil, fmt.Errorf("unable to unmarshal build plan: %w", err)
		}
	case app.RunnerJobGroupDeploy:
		err := json.Unmarshal([]byte(runnerPlan.PlanJSON), &compositePlan.DeployPlan)
		if err != nil {
			return nil, fmt.Errorf("unable to unmarshal deploy plan: %w", err)
		}
	case app.RunnerJobGroupActions:
		err := json.Unmarshal([]byte(runnerPlan.PlanJSON), &compositePlan.ActionPlan)
		if err != nil {
			return nil, fmt.Errorf("unable to unmarshal action plan: %w", err)
		}
	case app.RunnerJobGroupSandbox:
		err := json.Unmarshal([]byte(runnerPlan.PlanJSON), &compositePlan.SandboxPlan)
		if err != nil {
			return nil, fmt.Errorf("unable to unmarshal sandbox plan: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown runner job group: %s", runnerJob.Group)
	}

	return &compositePlan, nil
}
