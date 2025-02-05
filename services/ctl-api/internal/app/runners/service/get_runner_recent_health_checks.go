package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

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

	// Using raw SQL to query the CTE and return the result
	query := `
		WITH RankedRecords AS (
			SELECT 
				rhc.*, 
				toStartOfMinute(created_at) AS minute_bucket,
				ROW_NUMBER() OVER (PARTITION BY toStartOfMinute(created_at) ORDER BY rhc.created_at DESC) AS row_num
			FROM runner_health_checks AS rhc
			WHERE created_at >= NOW() - INTERVAL 1 HOUR
		)
		SELECT * FROM RankedRecords
		WHERE row_num = 1
		AND runner_id = ?
		ORDER BY created_at ASC
	`

	// Execute the query
	if err := s.chDB.Raw(query, runnerID).Scan(&healthChecks).Error; err != nil {
		return nil, fmt.Errorf("unable to get recent health checks: %w", err)
	}

	// Iterate through each health check record
	for _, healthCheck := range healthChecks {
		// Ensure the minute_bucket is set (in case it's not)
		if healthCheck.MinuteBucket.IsZero() {
			healthCheck.MinuteBucket = healthCheck.CreatedAt.Truncate(time.Minute)
		}
	}

	return healthChecks, nil
}
