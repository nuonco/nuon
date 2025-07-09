package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

type SyncSecretsRequest struct {
	PlanOnly bool `json:"plan_only"`
}

// @ID						SyncSecrets
// @Summary				sync secrets install
// @Description.markdown sync_secrets.md
// @Param					install_id	path	string							true	"install ID"
// @Param					req			body	SyncSecretsRequest	false	"Input"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{string}	ok
// @Router					/v1/installs/{install_id}/sync-secrets [post]
func (s *service) SyncSecrets(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req SyncSecretsRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	_, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	workflow, err := s.helpers.CreateInstallFlow(ctx,
		installID,
		app.WorkflowTypeSyncSecrets,
		map[string]string{},
		app.StepErrorBehaviorAbort,
		req.PlanOnly,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.evClient.Send(ctx, installID, &signals.Signal{
		Type:              signals.OperationExecuteFlow,
		InstallWorkflowID: workflow.ID,
	})

	ctx.Header(app.HeaderInstallWorkflowID, workflow.ID)

	ctx.JSON(http.StatusCreated, "ok")
}
