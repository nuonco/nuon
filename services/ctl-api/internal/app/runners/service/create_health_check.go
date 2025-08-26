package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateRunnerHealthCheckRequest struct {
	Process app.RunnerProcess `json:"process" swaggertype:"string"`
}

// @ID						CreateRunnerHealthCheck
// @Summary				create a runner health check
// @Description.markdown	create_runner_health_check.md
// @Param					req			body	CreateRunnerHealthCheckRequest	true	"Input"
// @Param					runner_id	path	string							true	"runner ID"
// @Tags					runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.RunnerHealthCheck
// @Router					/v1/runners/{runner_id}/health-checks [POST]
func (s *service) CreateRunnerHealthCheck(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req CreateRunnerHealthCheckRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	healthCheck, err := s.createRunnerHealthCheck(ctx, runnerID, req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create runner health check: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, healthCheck)
}

func (s *service) createRunnerHealthCheck(ctx context.Context, runnerID string, req CreateRunnerHealthCheckRequest) (*app.RunnerHealthCheck, error) {
	runnerHealthCheck := app.RunnerHealthCheck{
		RunnerID: runnerID,
	}
	if req.Process != "" {
		runnerHealthCheck.Process = req.Process
	} else {
		runnerHealthCheck.Process = app.RunnerProcessUknown
	}
	res := s.chDB.WithContext(ctx).
		Create(&runnerHealthCheck)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create runner health check: %w", res.Error)
	}

	return &runnerHealthCheck, nil
}
