package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetRunnerJobPlan
// @Summary				get runner job plan
// @Description.markdown	get_runner_job_plan.md
// @Param					runner_job_id	path	string	true	"runner job ID"
// @Tags					runners,runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	string
// @Router					/v1/runner-jobs/{runner_job_id}/plan [get]
func (s *service) GetRunnerJobPlan(ctx *gin.Context) {
	runnerJobID := ctx.Param("runner_job_id")

	plan, err := s.getRunnerJobPlan(ctx, runnerJobID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner job: %w", err))
		return
	}

	ctx.String(http.StatusOK, plan)
}

func (s *service) getRunnerJobPlan(ctx context.Context, runnerJobID string) (string, error) {
	var runnerPlan app.RunnerJobPlan

	res := s.db.WithContext(ctx).Where(app.RunnerJobPlan{
		RunnerJobID: runnerJobID,
	}).First(&runnerPlan)
	if res.Error != nil {
		return "", fmt.Errorf("unable to get job plan: %w", res.Error)
	}

	return runnerPlan.PlanJSON, nil
}
