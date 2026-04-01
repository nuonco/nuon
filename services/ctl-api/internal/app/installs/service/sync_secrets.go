package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
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
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	_, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	workflow, err := s.helpers.CreateWorkflow(ctx,
		installID,
		app.WorkflowTypeSyncSecrets,
		map[string]string{},
		req.PlanOnly,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	useQueues, err := s.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches, app.OrgFeatureQueues)
	if err != nil {
		ctx.Error(fmt.Errorf("checking features: %w", err))
		return
	}
	if useQueues {
		queueID, err := s.getInstallWorkflowsQueueID(ctx, installID)
		if err != nil {
			ctx.Error(err)
			return
		}
		if err := s.enqueueInstallSignal(ctx, queueID, &executeflow.Signal{
			InstallWorkflowID: workflow.ID,
		}); err != nil {
			ctx.Error(fmt.Errorf("enqueue signal: %w", err))
			return
		}
	} else {
		s.evClient.Send(ctx, installID, &signals.Signal{
			Type:              signals.OperationExecuteFlow,
			InstallWorkflowID: workflow.ID,
		})
	}

	ctx.Header(app.HeaderInstallWorkflowID, workflow.ID)

	ctx.JSON(http.StatusCreated, "ok")
}
