package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID GetRunnerJobPlan
// @Summary	get runner job plan
// @Description.markdown	get_runner_job_plan.md
// @Param			runner_job_id	path	string	true	"runner job ID"
// @Tags runners,runners/runner
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
// @Router			/v1/runner-jobs/{runner_job_id}/plan [get]
func (s *service) GetRunnerJobPlan(ctx *gin.Context) {
	runnerJobID := ctx.Param("runner_job_id")

	runnerJob, err := s.getRunnerJobPlan(ctx, runnerJobID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner job: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, runnerJob)
}

func (s *service) getRunnerJobPlan(ctx context.Context, runnerJobID string) (*planv1.Plan, error) {
	var runnerPlan app.RunnerJobPlan

	res := s.db.WithContext(ctx).Where(app.RunnerJobPlan{
		RunnerJobID: runnerJobID,
	}).First(&runnerPlan)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get job plan: %w", res.Error)
	}

	var plan planv1.Plan
	if err := proto.Unmarshal([]byte(runnerPlan.PlanJSON), &plan); err != nil {
		return nil, fmt.Errorf("unable to unmarshal job plan: %w", err)
	}

	return &plan, nil
}
