package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetRunnerRecentHealthChecks
// @Summary				get recent health checks
// @Description.markdown	get_runner_recent_health_checks.md
// @Param					runner_id					path	string	true	"runner ID"
// @Param					window						query	string	false	"window of health checks to return"	Default(1h)
// @Param					offset						query	int		false	"offset of results to return"		Default(0)
// @Param					limit						query	int		false	"limit of results to return"		Default(10)
// @Param					x-nuon-pagination-enabled	header	bool	false	"Enable pagination"
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.RunnerHealthCheck
// @Router					/v1/runners/{runner_id}/recent-health-checks [get]
func (s *service) GetRunnerRecentHealthChecks(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	_, err := s.getRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(err)
		return
	}

	windowStr := ctx.DefaultQuery("window", "1h")
	windowDur, err := time.ParseDuration(windowStr)
	if err != nil {
		ctx.Error(err)
		return
	}

	startTS := time.Now().Add(-windowDur)
	healthChecks, err := s.getRunnerRecentHealthChecks(ctx, runnerID, startTS)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, healthChecks)
}

func (s *service) getRunnerRecentHealthChecks(ctx *gin.Context, runnerID string, startTS time.Time) ([]*app.RunnerHealthCheck, error) {
	healthChecks := []*app.RunnerHealthCheck{}

	res := s.chDB.WithContext(ctx).
		Scopes(
			scopes.WithOverrideTable("runner_health_checks_view_v1"),
			scopes.WithOffsetPagination,
		).
		Where(app.RunnerHealthCheck{
			RunnerID: runnerID,
		}).
		Where("created_at > ?", startTS).
		Order("created_at asc").
		Find(&healthChecks)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get health checks")
	}

	healthChecks, err := db.HandlePaginatedResponse(ctx, healthChecks)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return healthChecks, nil
}
