package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID GetRunnerRecentHealthChecks
// @Summary	get recent health checks
// @Description.markdown get_runner_recent_health_checks.md
// @Param			runner_id	path	string	true	"runner ID"
// @Tags runners
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{array}	app.RunnerHealthCheck
// @Router			/v1/runners/{runner_id}/recent-health-checks [get]
func (s *service) GetRunnerRecentHealthChecks(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	_, err := s.getRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(err)
		return
	}

	healthChecks, err := s.getRunnerRecentHealthChecks(ctx, runnerID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, healthChecks)
}

func (s *service) getRunnerRecentHealthChecks(ctx context.Context, runnerID string) ([]*app.RunnerHealthCheck, error) {
	healthChecks := []*app.RunnerHealthCheck{}

	// last hour healthchecks
	resp := s.chDB.WithContext(ctx).
		Where("runner_id = ? AND created_at > NOW() - INTERVAL 60 MINUTE", runnerID).
		Order("created_at DESC").
		Find(&healthChecks)

	if resp.Error != nil {
		return nil, fmt.Errorf("failed to get recent health checks: %w", resp.Error)
	}

	return healthChecks, nil
}
