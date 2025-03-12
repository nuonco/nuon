package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

type AdminCreateHealthCheckJobRequest struct{}

// @ID AdminHealthCheckRunner
// @Summary	health check a runner
// @Description.markdown health_check_runner.md
// @Param			runner_id	path	string						true	"runner ID"
// @Param			req				body	AdminCreateHealthCheckJobRequest	true	"Input"
// @Tags runners/admin
// @Security AdminEmail
// @Accept			json
// @Produce		json
// @Success		201	{boolean}	true
// @Router			/v1/runners/{runner_id}/health-check-job [POST]
func (s *service) AdminCreateHealthCheck(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req AdminCreateHealthCheckJobRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	job, err := s.adminCreateAvailableJob(ctx, runnerID, app.RunnerJobTypeHealthCheck)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create health-check job: %w", err))
		return
	}

	s.evClient.Send(ctx, runnerID, &signals.Signal{
		Type:  signals.OperationHealthcheck,
		JobID: job.ID,
	})

	ctx.JSON(http.StatusCreated, true)
}

func (s *service) adminCreateAvailableJob(ctx context.Context, runnerID string, typ app.RunnerJobType) (*app.RunnerJob, error) {
	// identical to adminCreateJob but w/ staus ste to available
	// NOTE(fd): copied instead of extended since this is the only place we're using this atm
	runner, err := s.getRunner(ctx, runnerID)
	if err != nil {
		return nil, err
	}

	logStream := app.LogStream{
		OwnerID:   runner.ID,
		OwnerType: "runners",
		Open:      true,
		OrgID:     runner.OrgID,
	}
	if res := s.db.WithContext(ctx).Create(&logStream); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create log stream")
	}

	status := app.RunnerJobStatusAvailable
	runnerJob := app.RunnerJob{
		CreatedByID:       runner.CreatedByID,
		OrgID:             runner.OrgID,
		RunnerID:          runnerID,
		QueueTimeout:      time.Minute,
		ExecutionTimeout:  time.Second * 5,
		AvailableTimeout:  time.Second * 30,
		OverallTimeout:    time.Minute * 5,
		MaxExecutions:     1,
		Status:            status,
		Operation:         app.RunnerJobOperationTypeExec,
		StatusDescription: string(status),
		Type:              typ,
		Group:             app.RunnerJobGroupOperations,
		LogStreamID:       generics.ToPtr(logStream.ID),
	}
	if res := s.db.WithContext(ctx).Create(&runnerJob); res.Error != nil {
		return nil, res.Error
	}

	return &runnerJob, nil
}
