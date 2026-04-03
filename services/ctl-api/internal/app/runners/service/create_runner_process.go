package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type CreateRunnerProcessRequest struct {
	Type    app.RunnerProcessType `json:"type" validate:"required" swaggertype:"string"`
	Version string                `json:"version"`
}

// @ID						CreateRunnerProcess
// @Summary				create a runner process
// @Description.markdown	create_runner_process.md
// @Param					req			body	CreateRunnerProcessRequest	true	"Input"
// @Param					runner_id	path	string						true	"runner ID"
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
// @Success				201	{object}	app.RunnerProcess
// @Router					/v1/runners/{runner_id}/processes [POST]
func (s *service) CreateRunnerProcess(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req CreateRunnerProcessRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	process, err := s.createRunnerProcess(ctx, runnerID, req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create runner process: %w", err))
		return
	}

	if _, err := s.helpers.CreateProcessQueues(ctx, runnerID, process); err != nil {
		ctx.Error(fmt.Errorf("unable to create process queues: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, process)
}

func (s *service) createRunnerProcess(ctx context.Context, runnerID string, req CreateRunnerProcessRequest) (*app.RunnerProcess, error) {
	// create a log stream for this process
	logStream := app.LogStream{
		OwnerType: "runner_processes",
		Open:      true,
	}
	if res := s.db.WithContext(ctx).Create(&logStream); res.Error != nil {
		return nil, fmt.Errorf("unable to create log stream: %w", res.Error)
	}

	composite := app.NewCompositeStatus(ctx, app.Status(app.RunnerProcessStatusActive))
	composite.StatusHumanDescription = "This runner is still initializing and will not process jobs until its first health check"

	process := app.RunnerProcess{
		RunnerID:        runnerID,
		Type:            req.Type,
		Version:         req.Version,
		StartedAt:       generics.ToPtr(time.Now()),
		LogStreamID:     generics.ToPtr(logStream.ID),
		CompositeStatus: composite,
	}

	if res := s.db.WithContext(ctx).Create(&process); res.Error != nil {
		return nil, fmt.Errorf("unable to create runner process: %w", res.Error)
	}

	return &process, nil
}
