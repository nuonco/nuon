package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

type AdminCreateNoopJobRequest struct{}

// @ID AdminNoopRunner
// @Summary	trigger a noop runner job
// @Description.markdown create_noop_runner_job.md
// @Param			runner_id	path	string						true	"runner ID"
// @Param			req				body	AdminCreateNoopJobRequest	true	"Input"
// @Tags runners/admin
// @Accept			json
// @Produce		json
// @Success		201	{boolean}	true
// @Router			/v1/runners/{runner_id}/noop-job [POST]
func (s *service) AdminCreateNoopJob(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req AdminCreateNoopJobRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	job, err := s.adminCreateJob(ctx, runnerID, app.RunnerJobTypeNOOP)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create health check job: %w", err))
		return
	}

	s.evClient.Send(ctx, runnerID, &signals.Signal{
		Type:  signals.OperationJobQueued,
		JobID: job.ID,
	})

	ctx.JSON(http.StatusCreated, true)
}

func (s *service) adminCreateJob(ctx context.Context, runnerID string, typ app.RunnerJobType) (*app.RunnerJob, error) {
	runner, err := s.getRunner(ctx, runnerID)
	if err != nil {
		return nil, err
	}

	status := app.RunnerJobStatusQueued
	runnerJob := app.RunnerJob{
		CreatedByID:       runner.CreatedByID,
		OrgID:             runner.OrgID,
		RunnerID:          runnerID,
		QueueTimeout:      time.Minute,
		ExecutionTimeout:  time.Second * 5,
		AvailableTimeout:  time.Second * 30,
		OverallTimeout:    time.Minute * 5,
		MaxExecutions:     5,
		Status:            status,
		StatusDescription: string(status),
		Type:              typ,
		Group:             app.RunnerJobGroupOperations,
	}
	if res := s.db.WithContext(ctx).Create(&runnerJob); res.Error != nil {
		return nil, res.Error
	}

	return &runnerJob, nil
}
