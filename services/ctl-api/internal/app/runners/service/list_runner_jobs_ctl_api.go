package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						ListRunnerJobs
// @Summary				list org runner jobs
// @Description			list runner jobs for the current org that ran on the control plane. Used by orgs that build on the control plane and therefore have no org runner.
// @Param					group		query	string	false	"job group"
// @Param					groups		query	string	false	"job groups"
// @Param					status		query	string	false	"job status"
// @Param					statuses	query	string	false	"job statuses"
// @Param					executor	query	string	true	"job executor (must be control-plane)"
// @Param					offset		query	int		false	"offset of jobs to return"	Default(0)
// @Param					limit		query	int		false	"limit of jobs to return"	Default(10)
// @Param					page		query	int		false	"page number of results to return"	Default(0)
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.RunnerJob
// @Router					/v1/runner-jobs [get]
func (s *service) ListRunnerJobsCtlAPI(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	executor := app.RunnerJobExecutor(ctx.Query("executor"))
	if executor != app.RunnerJobExecutorControlPlane {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("unsupported executor %q", executor),
			Description: fmt.Sprintf("executor query param must be %q", app.RunnerJobExecutorControlPlane),
		})
		return
	}

	filters, err := parseRunnerJobFilters(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	runnerJobs, err := s.getRunnerJobsCtlAPI(ctx, org.ID, "", filters.statuses, filters.groups, executor, filters.limit)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, runnerJobs)
}
