package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
)

// @ID						AdminTriggerInstallStackOutputUpdate
// @Summary				trigger update install stack output for a run
// @Param					install_stack_version_run_id	path	string	true	"install stack version run ID"
// @Tags					installs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{boolean}	true
// @Router					/v1/install-stack-version-runs/{install_stack_version_run_id}/admin-trigger-stack-output-update [POST]
func (s *service) AdminTriggerInstallStackOutputUpdate(ctx *gin.Context) {
	runID := ctx.Param("install_stack_version_run_id")

	var run app.InstallStackVersionRun
	if res := s.db.WithContext(ctx).
		Preload("InstallStackVersion").
		Where("id = ?", runID).
		First(&run); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get install stack version run %s: %w", runID, res.Error))
		return
	}

	stackVersion := run.InstallStackVersion
	s.evClient.Send(ctx, stackVersion.InstallID, &signals.Signal{
		Type:           signals.OperationUpdateInstallStackOutputs,
		InstallStackID: stackVersion.InstallStackID,
	})
	ctx.JSON(http.StatusOK, true)
}
