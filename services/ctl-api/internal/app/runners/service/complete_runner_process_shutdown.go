package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @ID			CompleteRunnerProcessShutdown
// @Summary		hotfix stub: clears metadata.force_restart on the runner's group so the next /shutdowns poll returns empty
// @Param		runner_id	path	string	true	"runner ID"
// @Param		process_id	path	string	true	"process ID"
// @Param		shutdown_id	path	string	true	"shutdown ID"
// @Tags		runners/runner
// @Accept		json
// @Produce	json
// @Success	200	{object}	runnerProcessShutdownStub
// @Router		/v1/runners/{runner_id}/processes/{process_id}/shutdowns/{shutdown_id}/complete [POST]
func (s *service) CompleteRunnerProcessShutdown(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")
	processID := ctx.Param("process_id")
	shutdownID := ctx.Param("shutdown_id")

	s.db.WithContext(ctx).Exec(`
		UPDATE runner_group_settings
		SET metadata = metadata - 'force_restart'::text
		FROM runner_groups rg, runners r
		WHERE rg.id = runner_group_settings.runner_group_id
		  AND r.runner_group_id = rg.id
		  AND r.id = ?
	`, runnerID)

	ctx.JSON(http.StatusOK, runnerProcessShutdownStub{
		ID:              shutdownID,
		RunnerProcessID: processID,
		Type:            "graceful",
		Status:          "completed",
		CompositeStatus: map[string]any{"status": "completed"},
	})
}
