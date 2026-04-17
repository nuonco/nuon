package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *service) AdminListSandboxJobs(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var jobs []app.RunnerJob
	if res := s.db.WithContext(ctx).
		Where(app.RunnerJob{RunnerID: runnerID}).
		Order("created_at desc").
		Limit(50).
		Find(&jobs); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to list sandbox jobs: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, jobs)
}
