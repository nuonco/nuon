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

type AdminCreateJobRequest struct {
	Type app.RunnerJobType
}

// @ID AdminCreateJob
// @Summary	create a job
// @Description.markdown create_job.md
// @Param			runner_id	path	string						true	"runner ID"
// @Param			req				body	AdminCreateJobRequest	true	"Input"
// @Tags runners/admin
// @Accept			json
// @Produce		json
// @Success		200	{boolean}	true
// @Router			/v1/runners/{runner_id}/job [POST]
func (s *service) AdminCreateJob(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req AdminCreateJobRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	runnerJob, err := s.createRunnerJob(ctx, runnerID, req.Type)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create runner job: %w", err))
		return
	}

	s.evClient.Send(ctx, runnerID, &signals.Signal{
		Type:  signals.OperationJobQueued,
		JobID: runnerJob.ID,
	})
	ctx.JSON(http.StatusOK, true)
}

func (s *service) createRunnerJob(ctx context.Context, runnerID string, typ app.RunnerJobType) (*app.RunnerJob, error) {
	runnerJob := app.RunnerJob{
		RunnerID:          runnerID,
		QueueTimeout:      time.Second * 5,
		ExecutionTimeout:  time.Second * 5,
		OverallTimeout:    time.Second * 5,
		MaxExecutions:     5,
		Status:            app.RunnerJobStatusQueued,
		StatusDescription: string(app.RunnerJobStatusQueued),
	}
	if res := s.db.WithContext(ctx).Create(&runnerJob); res.Error != nil {
		return nil, res.Error
	}

	return &runnerJob, nil
}
