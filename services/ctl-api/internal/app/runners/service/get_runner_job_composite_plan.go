package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetRunnerJobCompositePlan
// @Summary				get runner job composite plan
// @Description.markdown	get_runner_job_composite_plan.md
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
// @Success				200	{object}	plantypes.CompositePlan
// @Router					/v1/runner-jobs/{runner_job_id}/composite-plan [get]
func (s *service) GetRunnerJobCompositePlan(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	runnerJobID := ctx.Param("runner_job_id")

	cp, err := s.getOrgRunnerJobCompositePlan(ctx, runnerJobID, org.ID)
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

	cp, err := runnerPlan.GetCompositePlan(blobstore.WithBlobService(ctx, s.blobSvc))
	if err != nil {
		return nil, err
	}
	if !cp.IsEmpty() {
		return cp, nil
	}

	// if empty derive from plan json

	var runnerJob app.RunnerJob
	res = s.db.WithContext(ctx).
		Where(app.RunnerJob{
			ID: runnerJobID,
		}).First(&runnerJob)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get job: %w", res.Error)
	}

	return runnerPlan.DeriveCompositePlan(&runnerJob)
}

func (s *service) getOrgRunnerJobCompositePlan(ctx context.Context, runnerJobID string, orgID string) (*plantypes.CompositePlan, error) {
	var runnerPlan app.RunnerJobPlan

	res := s.db.WithContext(ctx).
		Where("runner_job_id = ? AND org_id = ?", runnerJobID, orgID).
		First(&runnerPlan)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get job plan: %w", res.Error)
	}

	cp, err := runnerPlan.GetCompositePlan(blobstore.WithBlobService(ctx, s.blobSvc))
	if err != nil {
		return nil, err
	}
	if !cp.IsEmpty() {
		return cp, nil
	}

	// if empty derive from plan json

	var runnerJob app.RunnerJob
	res = s.db.WithContext(ctx).
		Where("id = ? AND org_id = ?", runnerJobID, orgID).
		First(&runnerJob)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get job: %w", res.Error)
	}

	return runnerPlan.DeriveCompositePlan(&runnerJob)
}
