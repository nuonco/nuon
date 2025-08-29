package service

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetRunnerJobs
// @Summary				get runner jobs
// @Description.markdown	get_runner_jobs.md
// @Param					group						query	string	false	"job group"
// @Param					groups						query	string	false	"job groups"
// @Param					status						query	string	false	"job status"
// @Param					statuses					query	string	false	"job statuses"
// @Param					runner_id					path	string	true	"runner ID"
// @Param					offset						query	int		false	"offset of jobs to return"	Default(0)
// @Param					limit						query	int		false	"limit of jobs to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
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
// @Success				200	{array}		app.RunnerJob
// @Router					/v1/runners/{runner_id}/jobs [get]
func (s *service) GetRunnerJobsCtlAPI(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	groups := []app.RunnerJobGroup{}
	groupStr := ctx.DefaultQuery("group", "")

	if groupStr != "" {
		groups = append(groups, app.RunnerJobGroup(groupStr))
	} else {
		groupsStr := ctx.DefaultQuery("groups", "")
		if groupsStr != "" {
			groupsStr := strings.Split(groupsStr, ",")
			// trim whitespace
			for i, groupStr := range groupsStr {
				groupsStr[i] = strings.TrimSpace(groupStr)
			}
			for _, groupStr := range groupsStr {
				groups = append(groups, app.RunnerJobGroup(groupStr))
			}
		}
	}

	statuses := []app.RunnerJobStatus{}
	statusStr := ctx.DefaultQuery("status", "")
	statusesStr := ctx.DefaultQuery("statuses", "")

	if statusStr != "" {
		status := app.RunnerJobStatus(statusStr)
		statuses = append(statuses, status)
	} else if statusesStr != "" {
		statusesStr := strings.Split(statusesStr, ",")
		// trim whitespace
		for i, statusStr := range statusesStr {
			statusesStr[i] = strings.TrimSpace(statusStr)
		}
		for _, statusStr := range statusesStr {
			statuses = append(statuses, app.RunnerJobStatus(statusStr))
		}
	}

	limitStr := ctx.DefaultQuery("limit", "60")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid limit %s: %w", limitStr, err),
			Description: "invalid limit",
		})
		return
	}

	runnerJobs, err := s.getRunnerJobsCtlAPI(ctx, runnerID, statuses, groups, limit)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, runnerJobs)
}

func (s *service) getRunnerJobsCtlAPI(ctx *gin.Context, runnerID string, statuses []app.RunnerJobStatus, groups []app.RunnerJobGroup, limit int) ([]*app.RunnerJob, error) {
	runnerJobs := []*app.RunnerJob{}

	where := app.RunnerJob{
		RunnerID: runnerID,
	}

	tx := s.db.WithContext(ctx).
		Limit(limit).
		Scopes(scopes.WithOffsetPagination).
		Where(where)

	if len(statuses) != 0 {
		tx = tx.Where("status in (?)", statuses)
	}

	if len(groups) != 0 {
		tx = tx.Where(`"group" in (?)`, groups)
	}

	res := tx.
		Scopes(scopes.WithDisableViews).
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
