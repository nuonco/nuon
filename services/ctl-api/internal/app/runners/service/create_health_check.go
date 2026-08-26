package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

type CreateRunnerHealthCheckRequest struct {
	Process app.RunnerProcessType `json:"process" swaggertype:"string"`
}

// @ID						CreateRunnerHealthCheck
// @Summary				create a runner health check
// @Description.markdown	create_runner_health_check.md
// @Param					req			body	CreateRunnerHealthCheckRequest	true	"Input"
// @Param					runner_id	path	string							true	"runner ID"
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
// @Success				201	{object}	app.RunnerHealthCheck
// @Router					/v1/runners/{runner_id}/health-checks [POST]
func (s *service) CreateRunnerHealthCheck(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req CreateRunnerHealthCheckRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	healthCheck, err := s.createRunnerHealthCheck(ctx, runnerID, req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create runner health check: %w", err))
		return
	}

	if trace.SpanFromContext(ctx.Request.Context()).IsRecording() {
		runner, err := s.heartbeatGetRunner(ctx, runnerID)
		if err != nil {
			s.l.Warn("unable to enrich runner health check span", zap.String("runner_id", runnerID), zap.Error(err))
		} else {
			var installID, installName string
			if runner.RunnerGroup.OwnerType == plugins.TableName(s.db, app.Install{}) {
				installID = runner.RunnerGroup.OwnerID
				if install := s.heartbeatGetInstall(ctx, installID); install != nil {
					installName = install.Name
				}
			}
			setRunnerSpanAttributes(ctx, runner, installID, installName,
				attribute.String("nuon.process_type", string(healthCheck.Process)),
			)
		}
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
		runnerHealthCheck.Process = app.RunnerProcessTypeUnknown
	}
	res := s.chDB.WithContext(ctx).
		Create(&runnerHealthCheck)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create runner health check: %w", res.Error)
	}

	return &runnerHealthCheck, nil
}
