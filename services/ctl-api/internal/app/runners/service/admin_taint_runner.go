package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @ID						AdminTaintRunner
// @Summary				taint a runner to exclude it from leader election
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Param					runner_id	path	string	true	"runner ID"
// @Success				200	{object}	app.Runner
// @Router					/v1/runners/{runner_id}/taint [post]
func (s *service) AdminTaintRunner(ctx *gin.Context) {
	runner, err := s.setRunnerTainted(ctx, ctx.Param("runner_id"), true, "")
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, runner)
}

// @ID						AdminUntaintRunner
// @Summary				untaint a runner to include it in leader election
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Param					runner_id	path	string	true	"runner ID"
// @Success				200	{object}	app.Runner
// @Router					/v1/runners/{runner_id}/untaint [post]
func (s *service) AdminUntaintRunner(ctx *gin.Context) {
	runner, err := s.setRunnerTainted(ctx, ctx.Param("runner_id"), false, "")
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, runner)
}
