package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

//	@ID						AdminGetRunnerJob
//	@Summary				get a runner job
//	@Description.markdown	admin_get_runner_job.md
//	@Param					runner_job_id	path	string	true	"runner ID"
//	@Tags					runners/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Produce				json
//	@Success				200	{object}	app.RunnerJob
//	@Router					/v1/runner-jobs/{runner_job_id} [GET]
func (s *service) AdminGetRunnerJob(ctx *gin.Context) {
	runnerJobID := ctx.Param("runner_job_id")

	runnerJob, err := s.adminGetRunnerJob(ctx, runnerJobID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner job: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, runnerJob)
}

func (s *service) adminGetRunnerJob(ctx context.Context, runnerJobID string) (*app.RunnerJob, error) {
	runnerJob := app.RunnerJob{}
	res := s.db.WithContext(ctx).
		Preload("Executions", func(db *gorm.DB) *gorm.DB {
			return db.Order("runner_job_executions.created_at DESC").Limit(1)
		}).
		Where("owner_id = ?", runnerJobID).
		Or("id = ?", runnerJobID).
		First(&runnerJob)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner job: %w", res.Error)
	}

	return &runnerJob, nil
}
