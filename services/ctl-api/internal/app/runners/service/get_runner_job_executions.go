package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetRunnerJobExecutions
// @Summary				get runner job executions
// @Description.markdown	get_runner_job_executions.md
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Param					runner_job_id				path	string	true	"runner job ID"
// @Tags					runners/runner,runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.RunnerJobExecution
// @Router					/v1/runner-jobs/{runner_job_id}/executions [get]
func (s *service) GetRunnerJobExecutions(ctx *gin.Context) {
	runnerJobID := ctx.Param("runner_job_id")

	runnerJobs, err := s.getRunnerJobExecutions(ctx, runnerJobID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner job executions: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, runnerJobs)
}

func (s *service) getRunnerJobExecutions(ctx *gin.Context, runnerJobID string) ([]app.RunnerJobExecution, error) {
	var runnerJob *app.RunnerJob
	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Preload("Executions", func(db *gorm.DB) *gorm.DB {
			return db.Order("runner_job_executions.created_at DESC").Limit(1000)
		}).
		First(&runnerJob, "id = ?", runnerJobID)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner job: %w", res.Error)
	}

	executions, err := db.HandlePaginatedResponse(ctx, runnerJob.Executions)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	runnerJob.Executions = executions

	return runnerJob.Executions, nil
}
