package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetRunnerJobs
// @Summary				get runner jobs
// @Description.markdown	get_runner_jobs.md
// @Param					runner_id					path	string	true	"runner ID"
// @Param					group						query	string	false	"job group"						Default(any)
// @Param					status						query	string	false	"job status"					Default(available)
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
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
// @Success				200	{array}		app.RunnerJob
// @Router					/v1/runners/{runner_id}/jobs [get]
func (s *service) GetRunnerJobs(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	groupStr := ctx.DefaultQuery("group", "any")
	grp := app.RunnerJobGroup(groupStr)

	limitStr := ctx.DefaultQuery("limit", "60")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid limit %s: %w", limitStr, err),
			Description: "invalid limit",
		})
		return
	}

	statusStr := ctx.DefaultQuery("status", "available")
	status := app.RunnerJobStatus(statusStr)

	runnerJobs, err := s.getRunnerJobs(ctx, runnerID, status, grp, limit)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, runnerJobs)
}

func (s *service) getRunnerJobs(ctx *gin.Context, runnerID string, status app.RunnerJobStatus, grp app.RunnerJobGroup, limit int) ([]*app.RunnerJob, error) {
	runnerJobs := []*app.RunnerJob{}

	where := app.RunnerJob{
		RunnerID: runnerID,
	}
	if status != app.RunnerJobStatusUnknown {
		where.Status = status
	}
	if grp != app.RunnerJobGroupAny {
		where.Group = grp
	}

	res := s.db.WithContext(ctx).
		Scopes(
			scopes.WithDisableViews,
			scopes.WithOffsetPagination,
		).
		Limit(limit).
		Where(where).
		Order("created_at desc").
		Find(&runnerJobs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner jobs: %w", res.Error)
	}

	runnerJobs, err := db.HandlePaginatedResponse(ctx, runnerJobs)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return runnerJobs, nil
}
