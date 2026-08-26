package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
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
// @Security				APIKey && OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
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

	var process *app.RunnerProcess
	err := traceRunnerRequestOperation(ctx, "runner.process.persist", func(ctx context.Context) error {
		var err error
		process, err = s.createRunnerProcess(ctx, runnerID, req)
		return err
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create runner process: %w", err))
		return
	}

	err = traceRunnerRequestOperation(ctx, "runner.process.queues.initialize", func(ctx context.Context) error {
		_, err := s.helpers.CreateProcessQueues(ctx, runnerID, process)
		return err
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create process queues: %w", err))
		return
	}

	_, telemetrySpan := runnerTracer.Start(
		ctx.Request.Context(), "runner.process.telemetry.emit")
	s.emitProcessStart(ctx, runnerID, process)
	telemetrySpan.End()

	ctx.JSON(http.StatusCreated, process)
}

func (s *service) emitProcessStart(ctx *gin.Context, runnerID string, process *app.RunnerProcess) {
	runner, err := s.heartbeatGetRunner(ctx, runnerID)
	if err != nil {
		s.l.Warn("unable to fetch runner for runner.process.start metric",
			zap.String("runner_id", runnerID),
			zap.Error(err),
		)
		return
	}

	var installID, installName string
	if runner.RunnerGroup.OwnerType == plugins.TableName(s.db, app.Install{}) {
		installID = runner.RunnerGroup.OwnerID
		if install := s.heartbeatGetInstall(ctx, installID); install != nil {
			installName = install.Name
		}
	}

	tagMap := make(map[string]string, 8)
	tagMap["org_id"] = runner.OrgID
	tagMap["org_name"] = runner.Org.Name
	tagMap["runner_id"] = runnerID
	tagMap["runner_type"] = string(runner.RunnerGroup.Type)
	tagMap["process_type"] = string(process.Type)
	tagMap["install_id"] = installID
	tagMap["install_name"] = installName
	tagMap["runner_version"] = process.Version

	s.mw.Incr("runner.process.start", metrics.ToTags(tagMap))

	setRunnerSpanAttributes(ctx, runner, installID, installName,
		attribute.String("nuon.runner_version", process.Version),
		attribute.String("nuon.process_id", process.ID),
		attribute.String("nuon.process_type", string(process.Type)),
	)
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

	composite := app.NewCompositeStatus(ctx, app.StatusPending)
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
