package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// @ID						BuildAppConfig
// @Summary				Build all components for an app config
// @Description			Creates a workflow that builds all components defined in the given app config.
// @Tags					apps
// @Accept					json
// @Param					app_id		path	string	true	"app ID"
// @Param					config_id	path	string	true	"app config ID"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.Workflow
// @Router					/v1/apps/{app_id}/configs/{config_id}/build [post]
func (s *service) BuildAppConfig(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	configID := ctx.Param("config_id")

	// Verify app exists and belongs to this org
	var a app.App
	res := s.db.WithContext(ctx).
		Where(app.App{OrgID: org.ID}).
		First(&a, "id = ?", appID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find app: %w", res.Error))
		return
	}

	// Verify config exists and belongs to this app
	var config app.AppConfig
	res = s.db.WithContext(ctx).
		Where(app.AppConfig{AppID: appID, OrgID: org.ID}).
		First(&config, "id = ?", configID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find app config: %w", res.Error))
		return
	}

	// Ensure the app has a queue for signal routing
	if err := s.helpers.EnsureAppQueue(ctx, appID); err != nil {
		ctx.Error(fmt.Errorf("unable to ensure app queue: %w", err))
		return
	}

	// Create the workflow (includes GenerateStepsSignal for shared flow infra)
	wf, err := s.helpers.CreateAppWorkflow(
		ctx,
		appID,
		app.WorkflowTypeAppConfigBuild,
		map[string]string{
			"app_config_id": configID,
		},
		false,
	)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create workflow: %w", err))
		return
	}

	// Find the app-workflows queue by name
	queue, err := s.queueClient.GetQueueByOwnerAndName(ctx, appID, plugins.TableName(s.db, app.App{}), appshelpers.AppWorkflowsQueueName)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to find app-workflows queue: %w", err))
		return
	}

	// Enqueue the shared execute-workflow signal to start the flow
	_, err = s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal: &executeflow.Signal{
			WorkflowID: wf.ID,
		},
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue build signal: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, wf)
}
