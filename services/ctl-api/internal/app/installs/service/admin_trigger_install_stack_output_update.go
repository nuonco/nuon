package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	updateinstallstackoutputs "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/updateinstallstackoutputs"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
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
	ctx2 := cctx.SetOrgIDContext(ctx, stackVersion.OrgID)

	useQueues, err := s.featuresClient.AllFeaturesEnabled(ctx2, app.OrgFeatureAppBranches, app.OrgFeatureQueues)
	if err != nil {
		ctx.Error(fmt.Errorf("checking features: %w", err))
		return
	}

	if useQueues {
		queueID, err := s.getInstallSignalsQueueID(ctx2, stackVersion.InstallID)
		if err != nil {
			ctx.Error(err)
			return
		}
		if err := s.enqueueInstallSignal(ctx2, queueID, &updateinstallstackoutputs.Signal{
			InstallStackID:        stackVersion.InstallStackID,
			InstallStackVersionID: run.InstallStackVersionID,
		}, "", ""); err != nil {
			ctx.Error(fmt.Errorf("enqueue signal: %w", err))
			return
		}
	} else {
		s.evClient.Send(ctx, stackVersion.InstallID, &signals.Signal{
			Type:                  signals.OperationUpdateInstallStackOutputs,
			InstallStackVersionID: run.InstallStackVersionID,
		})
	}

	ctx.JSON(http.StatusOK, true)
}
